package db

import (
	"log"
	"sort"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) UpdateUser(v models.User) error {

	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	v.JSON, err = v.MarshalJSON()
	if err != nil {
		return err
	}

	old := db.GetUser(v.ID)

	if old.BirthDate != v.BirthDate || old.Gender != v.Gender {
		userLocations := map[uint32]struct{}{}
		uv := db.GetUserVisits(v.ID)
		uv.M.RLock()
		for _, i := range uv.Visits {
			userLocations[i.Location] = struct{}{}
		}
		uv.M.RUnlock()
		for i := range userLocations {
			lm := db.GetLocationMarks(i)
			lm.M.Lock()
			for i := range lm.Marks {
				if lm.Marks[i].User == v.ID {
					lm.Marks[i].BirthDate = time.Unix(v.BirthDate, 0)
					lm.Marks[i].Gender = []byte(v.Gender)[0]
				}
			}
			lm.M.Unlock()
		}
	}

	db.lockU.Lock(v.ID)
	db.users[v.ID] = v
	db.lockU.Unlock(v.ID)

	return nil
}

func (db *DB) UpdateLocation(v models.Location) error {

	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	v.JSON, err = v.MarshalJSON()
	if err != nil {
		return err
	}

	old := db.GetLocation(v.ID)

	if old.Place != v.Place || old.Country != v.Country || old.Distance != v.Distance {
		locationUsers := map[uint32]struct{}{}
		lm := db.GetLocationMarks(v.ID)
		lm.M.RLock()
		for _, i := range lm.Marks {
			locationUsers[i.User] = struct{}{}
		}
		lm.M.RUnlock()
		for i := range locationUsers {
			uv := db.GetUserVisits(i)
			uv.M.Lock()
			for n, i := range uv.Visits {
				if i.Location == v.ID {
					i.Place = v.Place
					i.Country = v.Country
					i.Distance = v.Distance
					i.JSON, err = i.MarshalJSON()
					if err != nil {
						log.Fatal("UserVisit MarshalJSON failed:", err)
					}
					uv.Visits[n] = i
				}
			}
			uv.M.Unlock()
		}
	}

	db.lockL.Lock(v.ID)
	db.locations[v.ID] = v
	db.lockL.Unlock(v.ID)

	return nil
}

func (db *DB) UpdateVisit(v models.Visit) error {

	var err error
	err = v.Validate()
	if err != nil {
		return err
	}
	v.JSON, err = v.MarshalJSON()
	if err != nil {
		return err
	}

	old := db.GetVisit(v.ID)

	// move visit to new user
	if old.User != v.User {
		visit, found := db.GetUserVisits(old.User).Pop(v.ID)
		if !found {
			log.Fatalf("UserVisit of user %d for visit %d has been lost",
				old.User, v.ID)
		}
		db.GetUserVisits(v.User).Add(visit)
	}

	// move mark to new location
	if old.Location != v.Location {
		mark, found := db.GetLocationMarks(old.Location).Pop(v.ID)
		if !found {
			log.Fatalf("LocationMark of location %d for visit %d has been lost",
				old.Location, v.ID)
		}
		db.GetLocationMarks(v.Location).Add(mark)
	}

	user := db.GetUser(v.User)
	lm := db.GetLocationMarks(v.Location)
	lm.M.Lock()
	for n, i := range lm.Marks {
		if i.Visit == v.ID {
			lm.Marks[n] = models.LocationMark{
				Visit:     v.ID,
				User:      v.User,
				VisitedAt: v.VisitedAt,
				BirthDate: time.Unix(user.BirthDate, 0),
				Mark:      v.Mark,
				Gender:    []byte(user.Gender)[0],
			}
			break
		}
	}
	lm.M.Unlock()

	location := db.GetLocation(v.Location)
	uv := db.GetUserVisits(v.User)
	uv.M.Lock()
	for n, i := range uv.Visits {
		if i.Visit == v.ID {
			visit := models.UserVisit{
				Visit:     v.ID,
				Location:  v.Location,
				Mark:      v.Mark,
				VisitedAt: v.VisitedAt,
				Place:     location.Place,
				Country:   location.Country,
				Distance:  location.Distance,
			}
			visit.JSON, err = visit.MarshalJSON()
			if err != nil {
				log.Fatal("UserVisit MarshalJSON failed:", err)
			}
			uv.Visits[n] = visit
			break
		}
	}
	sort.Sort(models.UserVisitByVisitedAt(uv.Visits))
	uv.M.Unlock()

	db.lockV.Lock(v.ID)
	db.visits[v.ID] = v
	db.lockV.Unlock(v.ID)

	return nil
}
