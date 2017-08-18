// +build db_use_cmap

package models

//go:generate cmap-gen -package models -type User -key uint32
//go:generate cmap-gen -package models -type Location -key uint32
//go:generate cmap-gen -package models -type Visit -key uint32

//go:generate cmap-gen -package models -type *UserVisits -key uint32
//go:generate cmap-gen -package models -type *LocationMarks -key uint32
