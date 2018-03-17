package dao

import (
	"log"
	"gopkg.in/mgo.v2"
	 ."github.com/developinside3074/salas-go/models"
	"gopkg.in/mgo.v2/bson"
	"time"
	. "github.com/developinside3074/salas-go/models/enums"

)

type SalasDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "salas"
)

// Establecer coneccion con la base de datos
func (m *SalasDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// Recuperar la lista de Salas
func (m *SalasDAO) FindAll(id string) ([]Sala, error) {
	var salas []Sala

	err := db.C(COLLECTION).Find(bson.M{"idCentro": id}).Sort("nombre").All(&salas)
	return salas, err
}

// Recuperar una Sala por su identificador
func (m *SalasDAO) FindById(id string) (Sala, error) {
	var sala Sala
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&sala)
	return sala, err
}

// Insertar una Sala en la base de datos
func (m *SalasDAO) Insert(sala Sala) error {

	current := time.Now()

	err := db.C(COLLECTION).Insert(&Sala{ID: sala.ID,
	                                     IdCentro: sala.IdCentro,
	                                     IdZona: sala.IdZona,
	                                     IdArea: sala.IdArea,
	                                     IdDpto: sala.IdDpto,
	                                     IdServicio: sala.IdServicio,
	                                     Nombre: sala.Nombre,
									     Descripcion: sala.Descripcion,
									     Modalidades: sala.Modalidades,
									     UsuarioCreador: sala.UsuarioCreador,
									     UsuarioModificador: "",
									     FechaCreacion:  current.Format("2006-01-02 15:04:05"),
									     FechaModificacion: "",
									     Estado: ACTIVO})
	return err
}

// Eliminar una sala existente
func (m *SalasDAO) Delete(sala Sala) error {
	err := db.C(COLLECTION).Remove(&sala)
	return err
}

func (m *SalasDAO) Disable(sala Sala) error {

	sala.FechaModificacion = time.Now().Format("2006-01-02 15:04:05")
	sala.Estado = INACTIVO

	err := db.C(COLLECTION).UpdateId(sala.ID, &sala)
	return err
}

// Actualizar una sala
func (m *SalasDAO) Update(sala Sala) error {

	current := time.Now()

	sala.FechaModificacion = current.Format("2006-01-02 15:04:05")
	err := db.C(COLLECTION).UpdateId(sala.ID, &sala)
	return err
}
