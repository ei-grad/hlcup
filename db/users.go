package db

import (
	"fmt"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) GetUser(id uint32) models.User {
	return db.users.Get(id)
}

func (db *DB) AddUser(v models.User) error {
	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	_, err = db.sf.Do(fmt.Sprintf("user/%d", v.ID), func() (interface{}, error) {
		if db.users.Get(v.ID).IsValid() {
			return nil, fmt.Errorf("user %d already exist", v.ID)
		}
		db.users.Set(v.ID, v)
		return nil, nil
	})
	return err
}
