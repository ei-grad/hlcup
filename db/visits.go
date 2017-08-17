package db

import (
	"fmt"
	"time"

	"github.com/ei-grad/hlcup/models"
)

// GetVisit get visit by id
func (db *DB) GetVisit(id uint32) models.Visit {
	db.visitsMu.RLock()
	defer db.visitsMu.RUnlock()
	return db.visits.Get(id)
}

// AddVisit adds Visit to index
func (db *DB) AddVisit(v models.Visit) error {

	db.visitsMu.Lock()
	defer db.visitsMu.Unlock()

	if db.visits.Get(v.ID).IsValid() {
		return fmt.Errorf("visit with id %d already exists", v.ID)
	}
	db.visits.Set(v.ID, v)

	location := db.GetLocation(v.Location)
	if !location.Valid {
		return fmt.Errorf("location with id %d doesn't exist", v.Location)
	}

	user := db.GetUser(v.User)
	if !user.Valid {
		return fmt.Errorf("user with id %d doesn't exist", v.User)
	}

	// Add to db.locationMarks
	db.GetLocationMarks(v.Location).Add(models.LocationMark{
		Visit:     v.ID,
		User:      v.User,
		VisitedAt: v.VisitedAt,
		BirthDate: time.Unix(user.BirthDate, 0),
		Mark:      v.Mark,
		Gender:    []byte(user.Gender)[0],
	})

	// Add to db.userVisits
	db.GetUserVisits(v.User).Add(models.UserVisit{
		Visit:     v.ID,
		Location:  v.Location,
		Mark:      v.Mark,
		VisitedAt: v.VisitedAt,
		Place:     location.Place,
		Country:   location.Country,
		Distance:  location.Distance,
	})

	return nil

}
