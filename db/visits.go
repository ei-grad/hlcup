package db

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/ei-grad/hlcup/models"
)

type LocationMarkByVisitedAt []models.LocationMark

// Len is part of sort.Interface.
func (lm LocationMarkByVisitedAt) Len() int {
	return len(lm)
}

// Swap is part of sort.Interface.
func (lm LocationMarkByVisitedAt) Swap(i, j int) {
	lm[i], lm[j] = lm[j], lm[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (lm LocationMarkByVisitedAt) Less(i, j int) bool {
	return lm[i].VisitedAt < lm[j].VisitedAt
}

// GetVisit returns visit by its ID
func (db *DB) GetVisit(id uint32) models.Visit {
	return db.visits.Get(id)
}

// AddVisit adds Visit to index
func (db *DB) AddVisit(v models.Visit) error {

	db.visits.Set(v.ID, v)

	location := db.GetLocation(v.Location)
	if !location.Valid {
		log.Printf("location: %+v", location)
		log.Printf("visit: %+v", v)
		return fmt.Errorf("location with id %d doesn't exist", v.Location)
	}

	user := db.GetUser(v.User)
	if !user.Valid {
		return fmt.Errorf("user with id %d doesn't exist", v.User)
	}

	lm := db.locationMarks.Get(v.Location)
	if lm == nil {
		t, _ := db.locationSF.Do(strconv.Itoa(int(v.Location)), func() (interface{}, error) {
			ret := &models.LocationMarks{}
			db.locationMarks.Set(v.User, ret)
			return ret, nil
		})
		lm, _ = t.(*models.LocationMarks)
	}
	lm.M.Lock()
	lm.Marks = append(lm.Marks, models.LocationMark{
		VisitID:   v.ID,
		VisitedAt: v.VisitedAt,
		BirthDate: time.Unix(user.BirthDate, 0),
		Mark:      v.Mark,
		Gender:    []byte(user.Gender)[0],
	})
	sort.Sort(LocationMarkByVisitedAt(lm.Marks))
	lm.M.Unlock()

	uv := db.userVisits.Get(v.User)
	if uv == nil {
		t, _ := db.userSF.Do(strconv.Itoa(int(v.User)), func() (interface{}, error) {
			ret := &models.UserVisits{}
			db.userVisits.Set(v.User, ret)
			return ret, nil
		})
		uv, _ = t.(*models.UserVisits)
	}
	uv.M.Lock()
	uv.Visits = append(uv.Visits, models.UserVisit{
		Mark:      v.Mark,
		VisitedAt: v.VisitedAt,
		Place:     location.City,
		VisitID:   v.ID,
		Country:   location.Country,
		Distance:  location.Distance,
	})
	uv.M.Unlock()

	return nil

}

func (db *DB) GetLocationMarks(id uint32) (ret []models.LocationMark) {
	lm := db.locationMarks.Get(id)
	if lm != nil {
		ret = make([]models.LocationMark, len(lm.Marks))
		copy(ret, lm.Marks)
	}
	return
}

func (db *DB) GetUserVisits(id uint32) (ret []models.UserVisit) {
	uv := db.userVisits.Get(id)
	if uv != nil {
		ret = make([]models.UserVisit, len(uv.Visits))
		copy(ret, uv.Visits)
	}
	return
}
