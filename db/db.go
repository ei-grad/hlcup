package db

import (
	sf "github.com/golang/groupcache/singleflight"

	"github.com/ei-grad/hlcup/models"
)

type DB struct {
	users     *models.UserMap
	locations *models.LocationMap
	visits    *models.VisitMap

	locationMarks *models.LocationMarksMap
	userVisits    *models.UserVisitsMap

	locationSF sf.Group
	userSF     sf.Group
}

func New() *DB {
	return &DB{
		users:         models.NewUserMap(509),
		locations:     models.NewLocationMap(509),
		visits:        models.NewVisitMap(509),
		locationMarks: models.NewLocationMarksMap(509),
		userVisits:    models.NewUserVisitsMap(509),
	}
}

func (db *DB) GetUser(id uint32) models.User {
	return db.users.Get(id)
}

func (db *DB) AddUser(v models.User) {
	db.users.Set(v.ID, v)
}

func (db *DB) GetLocation(id uint32) models.Location {
	return db.locations.Get(id)
}

func (db *DB) AddLocation(v models.Location) {
	db.locations.Set(v.ID, v)
}

func (db *DB) GetVisit(id uint32) models.Visit {
	return db.visits.Get(id)
}
