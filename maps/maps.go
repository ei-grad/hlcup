package maps

import "github.com/ei-grad/hlcup/entities"

// User holds state for entities.User
//go:generate cmap-gen -package maps -type User
type User struct {
	Parsed entities.User
	JSON   []byte
}

// Location holds state for entities.Location
//go:generate cmap-gen -package maps -type Location
type Location struct {
	Parsed entities.Location
	JSON   []byte
}

// Visit holds state for entities.Visit
//go:generate cmap-gen -package maps -type Visit
type Visit struct {
	Parsed entities.Visit
	JSON   []byte
}
