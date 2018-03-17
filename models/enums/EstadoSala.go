package enums

type EstadoSala int16

const (
	ACTIVO   EstadoSala = 1 + iota
	INACTIVO
)
