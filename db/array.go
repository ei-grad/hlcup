package db

import (
	"errors"

	"github.com/ei-grad/hlcup/models"
)

const Version = "array"

const (
	MaxUsers           = 1500000
	MaxLocations       = 1100000
	MaxVisits          = 15000000
	DefaultShardsCount = 509
)

type DB struct {
	users     []models.User
	locations []models.Location
	visits    []models.Visit

	locationMarks []*models.LocationMarks
	userVisits    []*models.UserVisits

	lockU *ShardedLock
	lockL *ShardedLock
	lockV *ShardedLock

	lockLM *ShardedLock
	lockUV *ShardedLock
}

func New() *DB {

	db := new(DB)

	db.users = make([]models.User, MaxUsers)
	db.locations = make([]models.Location, MaxLocations)
	db.visits = make([]models.Visit, MaxVisits)

	db.locationMarks = make([]*models.LocationMarks, MaxLocations)
	db.userVisits = make([]*models.UserVisits, MaxUsers)

	db.lockU = NewShardedLock(DefaultShardsCount)
	db.lockL = NewShardedLock(DefaultShardsCount)
	db.lockV = NewShardedLock(DefaultShardsCount)
	db.lockLM = NewShardedLock(DefaultShardsCount)
	db.lockUV = NewShardedLock(DefaultShardsCount)

	return db
}

var ErrAlreadyExists = errors.New("already exists")

func (db *DB) GetUser(id uint32) models.User {
	if id >= MaxUsers {
		return models.User{}
	}
	db.lockU.RLock(id)
	defer db.lockU.RUnlock(id)
	return db.users[id]
}

func (db *DB) GetLocation(id uint32) models.Location {
	if id >= MaxLocations {
		return models.Location{}
	}
	db.lockL.RLock(id)
	defer db.lockL.RUnlock(id)
	return db.locations[id]
}

func (db *DB) GetVisit(id uint32) models.Visit {
	if id >= MaxVisits {
		return models.Visit{}
	}
	db.lockV.RLock(id)
	defer db.lockV.RUnlock(id)
	return db.visits[id]
}

func (db *DB) AddUser(v models.User) error {
	if err := v.Validate(); err != nil {
		return err
	}
	db.lockU.Lock(v.ID)
	if db.users[v.ID].IsValid() {
		return ErrAlreadyExists
	}
	db.users[v.ID] = v
	db.lockU.Unlock(v.ID)
	return nil
}

func (db *DB) AddLocation(v models.Location) error {
	if err := v.Validate(); err != nil {
		return err
	}
	db.lockL.Lock(v.ID)
	if db.locations[v.ID].IsValid() {
		return ErrAlreadyExists
	}
	db.locations[v.ID] = v
	db.lockL.Unlock(v.ID)
	return nil
}

func (db *DB) AddVisit(v models.Visit) error {
	if err := v.Validate(); err != nil {
		return err
	}
	db.lockV.Lock(v.ID)
	if db.visits[v.ID].IsValid() {
		return ErrAlreadyExists
	}
	db.visits[v.ID] = v
	db.lockV.Unlock(v.ID)
	return db.AddVisitToIndex(v)
}
