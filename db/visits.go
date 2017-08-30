package db

import (
	"fmt"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) AddVisitToIndex(v models.Visit) error {

	location := db.GetLocation(v.Location)
	if !location.IsValid() {
		return fmt.Errorf("location with id %d doesn't exist", v.Location)
	}

	user := db.GetUser(v.User)
	if !user.IsValid() {
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
	uv := models.UserVisit{
		Visit:     v.ID,
		Location:  v.Location,
		Mark:      v.Mark,
		VisitedAt: v.VisitedAt,
		Place:     location.Place,
		Country:   location.Country,
		Distance:  location.Distance,
	}
	db.GetUserVisits(v.User).Add(uv)

	return nil

}
