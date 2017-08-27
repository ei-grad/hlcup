package db

import (
	"errors"

	"github.com/ei-grad/hlcup/models"
)

const Version = "array"

const (
	MaxUsers           = 2000000
	MaxLocations       = 1000000
	MaxVisits          = MaxUsers * 10
	DefaultShardsCount = 41 // it is a prime near 42
)

type DB struct {
	users     [MaxUsers]models.User
	locations [MaxLocations]models.Location
	visits    [MaxVisits]models.Visit

	locationMarks [MaxLocations]*models.LocationMarks
	userVisits    [MaxUsers]*models.UserVisits

	lockU *ShardedLock
	lockL *ShardedLock
	lockV *ShardedLock

	lockLM *ShardedLock
	lockUV *ShardedLock
}

func New() *DB {
	return &DB{
		lockU: NewShardedLock(DefaultShardsCount),
		lockL: NewShardedLock(DefaultShardsCount),
		lockV: NewShardedLock(DefaultShardsCount),

		lockLM: NewShardedLock(DefaultShardsCount),
		lockUV: NewShardedLock(DefaultShardsCount),
	}
}

var ErrAlreadyExists = errors.New("already exists")

func (db *DB) GetUser(id uint32) models.User {
	db.lockU.RLock(id)
	defer db.lockU.RUnlock(id)
	return db.users[id]
}

func (db *DB) GetLocation(id uint32) models.Location {
	db.lockL.RLock(id)
	defer db.lockL.RUnlock(id)
	return db.locations[id]
}

func (db *DB) GetVisit(id uint32) models.Visit {
	db.lockV.RLock(id)
	defer db.lockV.RUnlock(id)
	return db.visits[id]
}

func (db *DB) AddUser(v models.User) error {
	if err := v.Validate(); err != nil {
		return err
	}
	v.JSON, _ = v.MarshalJSON()
	db.lockU.Lock(v.ID)
	if db.users[v.ID].Valid {
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
	v.JSON, _ = v.MarshalJSON()
	db.lockL.Lock(v.ID)
	if db.locations[v.ID].Valid {
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
	v.JSON, _ = v.MarshalJSON()
	db.lockV.Lock(v.ID)
	if db.visits[v.ID].Valid {
		return ErrAlreadyExists
	}
	db.visits[v.ID] = v
	db.lockV.Unlock(v.ID)
	return db.AddVisitToIndex(v)
}
