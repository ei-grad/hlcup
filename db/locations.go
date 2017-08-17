package db

import (
	"fmt"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) GetLocation(id uint32) models.Location {
	return db.locations.Get(id)
}

func (db *DB) AddLocation(v models.Location) error {
	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	_, err = db.sf.Do(fmt.Sprintf("location/%d", v.ID), func() (interface{}, error) {
		if db.locations.Get(v.ID).IsValid() {
			return nil, fmt.Errorf("location %d already exist", v.ID)
		}
		db.locations.Set(v.ID, v)
		return nil, nil
	})
	return err
}
