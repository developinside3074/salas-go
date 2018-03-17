package main

import (
	"github.com/gorilla/mux"
	"encoding/json"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"log"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	."github.com/developinside3074/salas-go/config"
	."github.com/developinside3074/salas-go/dao"
	."github.com/developinside3074/salas-go/models"
	"github.com/developinside3074/salas-go/eureka" //Paquete externo
	"fmt"
)

var config = Config{}
var dao = SalasDAO{}

const (
	NombreApp = "msSalasGo"
	EurekaURL = "http://172.104.35.23:8761"
	PuertoApp = "8096"
	PuertoSeguro = "443"
)

// GET list of movies
func AllSalasEndPoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	salas, err := dao.FindAll(params["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, salas)
}

// Obtener una Sala por su ID
func FindSalasByIdEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sala, err := dao.FindById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID invalido para Sala")
		return
	}
	respondWithJson(w, http.StatusOK, sala)
}

// POST para crear una nueva Sala
func CreateSalasEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var sala Sala
	if err := json.NewDecoder(r.Body).Decode(&sala); err != nil {
		respondWithError(w, http.StatusBadRequest, "Petición de solicitud inválida")
		return
	}

	sala.ID = bson.NewObjectId()

	if err := dao.Insert(sala); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//Recuperar JSON de la base de datos
	salaToFound, err := dao.FindById(sala.ID.Hex())
	log.Println("Sala: {}", salaToFound)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID invalido para Sala")
		return
	}
	respondWithJson(w, http.StatusOK, salaToFound)

}

// PUT update an existing movie
func UpdateSalasEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var sala Sala
	if err := json.NewDecoder(r.Body).Decode(&sala); err != nil {
		respondWithError(w, http.StatusBadRequest, "Petición de solicitud inválida")
		return
	}
	if err := dao.Update(sala); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//Recuperar JSON de la base de datos
	salaToFound, err := dao.FindById(sala.ID.Hex())

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID invalido para Sala")
		return
	}
	respondWithJson(w, http.StatusOK, salaToFound)
}

// Inhabilitar una Sala Existente
func DisableSalasEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var sala Sala
	if err := json.NewDecoder(r.Body).Decode(&sala); err != nil {
		respondWithError(w, http.StatusBadRequest, "ID invalido para Sala")
		return
	}
	if err := dao.Disable(sala); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//Recuperar JSON de la base de datos
	salaToFound, err := dao.FindById(sala.ID.Hex())

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID invalido para Sala")
		return
	}
	respondWithJson(w, http.StatusOK, salaToFound)

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Parse the configuration file 'config.toml', and establish a connection to DB
func init() {
	config.Read()
	dao.Server = config.Server
	dao.Database = config.Database

	dao.Connect()


}

// Define HTTP request routes
func main() {

	fmt.Println("Nombre de aplicacion",NombreApp)
	fmt.Println("Puerto de app",PuertoApp)
	fmt.Println("eurekaURL: ", EurekaURL)

	handleSigterm()                              // Asegura el cierre sobre Ctrl+C o kill

	go startWebServer()                          // Inicializa HTTP servicio  (async)

	eureka.RegisterAt(EurekaURL, NombreApp, PuertoApp,PuertoSeguro) // Realiza el registro de Eureka

	go eureka.StartHeartbeat(NombreApp)   // Realiza latencia Eureka Server (async)

	// Block...
	wg := sync.WaitGroup{}                       // Uso de WaitGroup para bloquear main() y salir
	wg.Add(2)
	wg.Wait()

}

func handleSigterm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		eureka.Deregister(NombreApp)
		os.Exit(1)
	}()
}

func startWebServer() {
	log.Println("Iniciando HTTP escuchando en el puerto", PuertoApp)
	r := mux.NewRouter()
	r.HandleFunc("/v1/centrosAsistenciales/{id}/salas/todos", AllSalasEndPoint).Methods("GET")
	r.HandleFunc("/v1/salas", CreateSalasEndPoint).Methods("POST")
	r.HandleFunc("/v1/salas", UpdateSalasEndPoint).Methods("PUT")
	r.HandleFunc("/v1/salas", DisableSalasEndPoint).Methods("DELETE")
	//r.HandleFunc("/v1/salas/{id}", DeleteSalasEndPoint).Methods("DELETE")
	r.HandleFunc("/v1/salas/{id}", FindSalasByIdEndpoint).Methods("GET")
	if err := http.ListenAndServe(":" + PuertoApp , r); err != nil {
		log.Println("Ha ocurrido un error HTTP en el puerto receptor ", PuertoApp)
		log.Println("Error: ", err.Error())
	}
}