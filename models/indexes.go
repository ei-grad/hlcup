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
//ffjson:skip
type LocationMark struct {
	VisitID   uint32
	VisitedAt int
	BirthDate time.Time
	Gender    rune
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

	VisitID  uint32 `json:"-"`
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
