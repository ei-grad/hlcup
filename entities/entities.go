package entities

import "bytes"

type Entity int

const (
	Unknown Entity = iota
	User
	Location
	Visit
)

var (
	bytesUsers     = []byte("users")
	bytesLocations = []byte("locations")
	bytesVisits    = []byte("visits")
)

func GetEntityByRoute(entity []byte) Entity {
	switch {
	case bytes.Equal(entity, bytesUsers):
		return User
	case bytes.Equal(entity, bytesLocations):
		return Location
	case bytes.Equal(entity, bytesVisits):
		return Visit
	default:
		return Unknown
	}
}

func GetEntityRoute(entity Entity) []byte {
	switch entity {
	case User:
		return bytesUsers
	case Visit:
		return bytesVisits
	case Location:
		return bytesLocations
	default:
		panic("Unknown entity in GetEntityRoute")
	}
}
