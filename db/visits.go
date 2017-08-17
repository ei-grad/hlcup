package db

import (
	"fmt"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) GetVisit(id uint32) models.Visit {
	return db.visits.Get(id)
}

func (db *DB) AddVisit(v models.Visit) error {
	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	_, err = db.sf.Do(fmt.Sprintf("visit/%d", v.ID), func() (interface{}, error) {
		if db.visits.Get(v.ID).IsValid() {
			return nil, fmt.Errorf("visit %d already exist", v.ID)
		}
		db.visits.Set(v.ID, v)
		return nil, nil
	})
	if err != nil {
		return err
	}

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
