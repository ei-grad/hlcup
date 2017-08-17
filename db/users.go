package db

import (
	"fmt"

	"github.com/ei-grad/hlcup/models"
)

// GetUser get user by id
func (db *DB) GetUser(id uint32) models.User {
	return db.users.Get(id)
}

// AddUser adds user to database
func (db *DB) AddUser(v models.User) error {
	if db.users.Get(v.ID).IsValid() {
		return fmt.Errorf("user %d already exist", v.ID)
	}
	db.users.Set(v.ID, v)
	return nil
}
