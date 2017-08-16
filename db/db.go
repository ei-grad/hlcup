package db

import (
	sf "github.com/golang/groupcache/singleflight"
	"sync"

	"github.com/ei-grad/hlcup/models"
)

type Users []models.User

func (s Users) Get(id uint32) models.User {
	return s[id]
}

func (s Users) Set(id uint32, v models.User) {
	s[id] = v
}

type Locations []models.Location

func (s Locations) Get(id uint32) models.Location {
	return s[id]
}

func (s Locations) Set(id uint32, v models.Location) {
	s[id] = v
}

type Visits []models.Visit

func (s Visits) Get(id uint32) models.Visit {
	return s[id]
}

func (s Visits) Set(id uint32, v models.Visit) {
	s[id] = v
}

type LocationMarks []*models.LocationMarks

func (s LocationMarks) Get(id uint32) *models.LocationMarks {
	return s[id]
}

func (s LocationMarks) Set(id uint32, v *models.LocationMarks) {
	s[id] = v
}

type UserVisits []*models.UserVisits

func (s UserVisits) Get(id uint32) *models.UserVisits {
	return s[id]
}

func (s UserVisits) Set(id uint32, v *models.UserVisits) {
	s[id] = v
}

// DB is inmemory database optimized for its task
type DB struct {
	users     Users
	locations Locations
	visits    Visits

	locationMarks LocationMarks
	userVisits    UserVisits

	locationSF sf.Group
	userSF     sf.Group

	mu sync.RWMutex
}

var (
	maxUsers     = 500000
	maxLocations = 500000
	maxVisits    = 5000000
)

// New creates new DB
func New() *DB {
	return &DB{
		users:         make([]models.User, maxUsers),
		locations:     make([]models.Location, maxLocations),
		visits:        make([]models.Visit, maxVisits),
		locationMarks: make([]*models.LocationMarks, maxLocations),
		userVisits:    make([]*models.UserVisits, maxUsers),
	}
}
