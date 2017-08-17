package db

import (
	"fmt"

	"github.com/ei-grad/hlcup/models"
)

// GetLocation get location by id
func (db *DB) GetLocation(id uint32) models.Location {
	db.locationsMu.RLock()
	defer db.locationsMu.RUnlock()
	return db.locations.Get(id)
}

// AddLocation add location to database
func (db *DB) AddLocation(v models.Location) error {
	db.locationsMu.Lock()
	defer db.locationsMu.Unlock()
	if db.locations.Get(v.ID).IsValid() {
		return fmt.Errorf("location %d already exist", v.ID)
	}
	db.locations.Set(v.ID, v)
	return nil
}
