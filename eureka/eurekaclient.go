package eureka
/**
The MIT License (MIT)
Copyright (c) 2016 ErikL
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
import (
	"fmt"
	"net"
	"github.com/twinj/uuid"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"
	"log"
	"encoding/json"
)

var instanceId string
var discoveryServerUrl = "http://localhost:8761"

var regTpl = `{
  "instance": {
    "hostName":"${ipAddress}",
    "app":"${appName}",
    "ipAddr":"${ipAddress}",
    "vipAddress":"${appName}",
    "status":"UP",
    "port": {
      "$":${port},
      "@enabled": true
    },
    "securePort": {
      "$":${securePort},
      "@enabled": true
    },
    "homePageUrl" : "http://${ipAddress}:${port}/",
    "statusPageUrl": "http://${ipAddress}:${port}/info",
    "healthCheckUrl": "http://${ipAddress}:${port}/health",
    "dataCenterInfo" : {
      "@class":"com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
      "name": "MyOwn"
    },
    "metadata": {
      "instanceId" : "${appName}:${instanceId}"
    }
  }
}`

/**
 * Registre la aplicación en eurekaUrl por defecto.
 */
func RegisterAt(eurekaUrl string, appName string, port string, securePort string) {
	discoveryServerUrl = eurekaUrl
	Register(appName, port, securePort)
}

/**
 * Registre la aplicación en eurekaUrl por defecto eurekaUrl.
 */
func Register(appName string, port string, securePort string) {
	instanceId = getUUID()

	tpl := string(regTpl)
	tpl = strings.Replace(tpl, "${ipAddress}", getLocalIP(), -1)
	tpl = strings.Replace(tpl, "${port}", port, -1)
	tpl = strings.Replace(tpl, "${securePort}", securePort, -1)
	tpl = strings.Replace(tpl, "${instanceId}", instanceId, -1)
	tpl = strings.Replace(tpl, "${appName}", appName, -1)

	// Registro.
	registerAction := HttpAction{
		Url:         discoveryServerUrl + "/eureka/apps/" + appName,
		Method:      "POST",
		ContentType: "application/json;charset=UTF-8",
		Body:        tpl,
	}
	var result bool
	for {
		result = doHttpRequest(registerAction)
		if result {
			fmt.Println("Registro Exitoso OK")
			handleSigterm(appName)
			go StartHeartbeat(appName)
			break
		} else {
			fmt.Println("Intento de registro de " + appName + " fallido...")
			time.Sleep(time.Second * 5)
		}
	}

}

/**
 * Dado el appName proporcionado, este func consulta la API de Eureka para las instancias de appName y devuelve
 * como una estructura EurekaApplication.
 */
func GetServiceInstances(appName string) ([]EurekaInstance, error) {
	var m EurekaServiceResponse
	fmt.Println("Consultando a Eureka por la instacia de " + appName + " en: " + discoveryServerUrl + "/eureka/apps/" + appName)
	queryAction := HttpAction{
		Url:         discoveryServerUrl + "/eureka/apps/" + appName,
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	log.Println("Haciendo consulta usando la URL: " + queryAction.Url)
	bytes, err := executeQuery(queryAction)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("Respuesta conseguida de Eureka:\n" + string(bytes))
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			fmt.Println("Problema para parsear JSON respuesta de Eureka: " + err.Error())
			return nil, err
		}
		return m.Application.Instance, nil
	}
}

// Experimental, no probado.
func GetServices() ([]EurekaApplication, error) {
	var m EurekaApplicationsRootResponse
	fmt.Println("Consultando a eureka por servicios en: " + discoveryServerUrl + "/eureka/apps")
	queryAction := HttpAction{
		Url:         discoveryServerUrl + "/eureka/apps",
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	log.Println("Haciendo consulta usando la URL: " + queryAction.Url)
	bytes, err := executeQuery(queryAction)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("Servicios obtenidos como respuesta de Eureka:\n" + string(bytes))
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			fmt.Println("Problema para parsear JSON respuesta de Eureka:: " + err.Error())
			return nil, err
		}
		return m.Resp.Applications, nil
	}
}

// Iniciar una go rutina, se repetirá indefinidamente hasta que la aplicación salga
func StartHeartbeat(appName string) {
	for {
		time.Sleep(time.Second * 30)
		heartbeat(appName)
	}
}

func heartbeat(appName string) {
	heartbeatAction := HttpAction{
		Url:         discoveryServerUrl + "/eureka/apps/" + appName + "/" + getLocalIP() + ":" + appName + ":" + instanceId,
		Method:      "PUT",
		ContentType: "application/json;charset=UTF-8",
	}
	fmt.Println("Emitir latidos a " + heartbeatAction.Url)
	doHttpRequest(heartbeatAction)
}

// Desregistrar al servisio del servidor Eureka
func Deregister(appName string) {
	fmt.Println("Intentando eliminar el registro de la aplicación " + appName + "...")
	// Desregistrar
	deregisterAction := HttpAction{
		Url:         discoveryServerUrl + "/eureka/apps/" + appName + "/" + getLocalIP() + ":" + appName + ":" + instanceId,
		ContentType: "application/json;charset=UTF-8",
		Method:      "DELETE",
	}
	doHttpRequest(deregisterAction)
	fmt.Println("Desregistrado aplicacion " + appName + ", existente. Revisado por Eureka...")
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// Verifica el tipo de dirección y si no es un loopback, muéstralo
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("No se puede determinar la dirección IP local (sin loopback). Saliendo")
}

func getUUID() string {
	return uuid.NewV4().String()
}

func handleSigterm(appName string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		Deregister(appName)
		os.Exit(1)
	}()
}
