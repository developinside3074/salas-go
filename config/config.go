package config

import (
	"log"
	"github.com/BurntSushi/toml"
)

// Representa las credenciales de la base de datos
type Config struct {
	Server       string
	Database     string
}

// Lee y parsea la configuracion
func (c *Config) Read() {
	if _, err := toml.DecodeFile("config.toml", &c); err != nil {
		log.Fatal(err)
	}
}
