// +build !db_use_cmap

package db

import (
	"github.com/golang/groupcache/singleflight"
	"sync"

	"github.com/ei-grad/hlcup/models"
)

type Users struct {
	mu   sync.RWMutex
	data [1000000]models.User
}

func (s *Users) Get(id uint32) models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *Users) Set(id uint32, v models.User) {
	s.mu.Lock()
	s.data[id] = v
	s.mu.Unlock()
}

type Locations struct {
	mu   sync.RWMutex
	data [1000000]models.Location
}

func (s *Locations) Get(id uint32) models.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *Locations) Set(id uint32, v models.Location) {
	s.mu.Lock()
	s.data[id] = v
	s.mu.Unlock()
}

type Visits struct {
	mu   sync.RWMutex
	data [10000000]models.Visit
}

func (s *Visits) Get(id uint32) models.Visit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *Visits) Set(id uint32, v models.Visit) {
	s.mu.Lock()
	s.data[id] = v
	s.mu.Unlock()
}

type LocationMarks struct {
	mu   sync.RWMutex
	data [1000000]*models.LocationMarks
}

func (s *LocationMarks) Get(id uint32) *models.LocationMarks {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *LocationMarks) Set(id uint32, v *models.LocationMarks) {
	s.mu.Lock()
	s.data[id] = v
	s.mu.Unlock()
}

type UserVisits struct {
	mu   sync.RWMutex
	data [1000000]*models.UserVisits
}

func (s *UserVisits) Get(id uint32) *models.UserVisits {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *UserVisits) Set(id uint32, v *models.UserVisits) {
	s.mu.Lock()
	s.data[id] = v
	s.mu.Unlock()
}

// DB is inmemory database optimized for its task
type DB struct {
	users     Users
	locations Locations
	visits    Visits

	locationMarks LocationMarks
	userVisits    UserVisits

	sf singleflight.Group
}

// New creates new DB
func New() *DB {
	return &DB{}
}
