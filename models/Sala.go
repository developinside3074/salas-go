package models

import (
	"gopkg.in/mgo.v2/bson"
	. "github.com/developinside3074/salas-go/models/enums"
)

type Sala struct {
	ID                 bson.ObjectId `bson:"_id" json:"id"`
	IdCentro           string        `bson:"idCentro" json:"idCentro"`
	IdZona             string        `bson:"idZona" json:"idZona"`
	IdArea             string        `bson:"idArea" json:"idArea"`
	IdDpto             string        `bson:"idDpto" json:"idDpto"`
	IdServicio         string        `bson:"idServicio" json:"idServicio"`
	Nombre             string        `bson:"nombre" json:"nombre"`
	Descripcion        string        `bson:"descripcion" json:"descripcion"`
	Modalidades        []string      `bson:"modalidades" json:"modalidades"`
	UsuarioCreador     string        `bson:"usuarioCreador" json:"usuarioCreador"`
	UsuarioModificador string        `bson:"usuarioModificador" json:"usuarioModificador"`
	FechaCreacion      string        `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaModificacion  string     	 `bson:"fechaModificacion" json:"fechaModificacion"`
	Estado             EstadoSala    `bson:"estado" json:"estado"`
}
