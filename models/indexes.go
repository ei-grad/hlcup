package models

import (
	"sync"
	"time"
)

//go:generate ffjson $GOFILE

// LocationMark contains info needed to implement filters:
//    fromDate - учитывать оценки только с visited_at > fromDate
//    toDate - учитывать оценки только с visited_at < toDate
//    fromAge - учитывать только путешественников, у которых возраст (считается от текущего timestamp) больше этого параметра
//    toAge - как предыдущее, но наоборот
//    gender - учитывать оценки только мужчин или женщин
// ffjson используется для отладочной ручки /location/<id>/marks
type LocationMark struct {
	Visit     uint32
	User      uint32
	VisitedAt int
	BirthDate time.Time
	Gender    byte
	Mark      uint8
}

// LocationMarks is used to calculate average location mark
//go:generate cmap-gen -package models -type *LocationMarks -key uint32
//ffjson:skip
type LocationMarks struct {
	M     sync.RWMutex
	Marks []LocationMark
}

// UserVisit is used to filter and output the user visit info
//    fromDate - посещения с visited_at > fromDate
//    toDate - посещения с visited_at < toDate
//    country - название страны, в которой находятся интересующие достопримечательности
//    toDistance - возвращать только те места, у которых расстояние от города меньше этого параметра
type UserVisit struct {
	Mark      uint8  `json:"mark"`
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`

	Visit    uint32 `json:"-"`
	Location uint32 `json:"-"`
	Country  string `json:"-"`
	Distance uint32 `json:"-"`
}

// UserVisits is user visits index
//go:generate cmap-gen -package models -type *UserVisits -key uint32
//ffjson:skip
type UserVisits struct {
	M      sync.RWMutex
	Visits []UserVisit
}

// UserLocations holds a list of location IDs which the user has visited
//go:generate cmap-gen -package models -type *UserLocations -key uint32
//ffjson:skip
type UserLocations struct {
	M         sync.RWMutex
	Locations []uint32
}

// LocationUsers holds a list of users IDs which visited the location
//go:generate cmap-gen -package models -type *LocationUsers -key uint32
//ffjson:skip
type LocationUsers struct {
	M     sync.RWMutex
	Users []uint32
}
