package db

import (
	"github.com/ei-grad/hlcup/models"
)

// TODO: update LocationMarks and UserVisits

func (db *DB) UpdateUser(v models.User) {
	db.users.Set(v.ID, v)
}

func (db *DB) UpdateLocation(v models.Location) {
	db.locations.Set(v.ID, v)
}

func (db *DB) UpdateVisit(v models.Visit) {
	db.visits.Set(v.ID, v)
}
