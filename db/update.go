package db

import (
	"fmt"
	"sort"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) UpdateUser(v models.User) {

	old := db.users.Get(v.ID)

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
			for _, i := range lm.Marks {
				if i.User == v.ID {
					i.BirthDate = time.Unix(v.BirthDate, 0)
					i.Gender = []byte(v.Gender)[0]
				}
			}
			lm.M.Unlock()
		}
	}

	db.users.Set(v.ID, v)
}

func (db *DB) UpdateLocation(v models.Location) {

	old := db.locations.Get(v.ID)

	if old.Place != v.Place || old.Country != v.Country || old.Distance != v.Distance {
		locationUsers := map[uint32]struct{}{}
		lm := db.GetLocationMarks(v.ID)
		lm.M.RLock()
		for _, i := range lm.Marks {
			locationUsers[i.User] = struct{}{}
		}
		lm.M.RUnlock()
		for i := range locationUsers {
			uv := db.userVisits.Get(i)
			if uv == nil {
				continue
			}
			uv.M.Lock()
			for _, i := range uv.Visits {
				if i.Location == v.ID {
					i.Place = v.Place
					i.Country = v.Country
					i.Distance = v.Distance
				}
			}
			uv.M.Unlock()
		}
	}

	db.locations.Set(v.ID, v)
}

func (db *DB) UpdateVisit(v models.Visit) {

	old := db.visits.Get(v.ID)

	// move visit to new user
	if old.User != v.User {
		visit, found := db.GetUserVisits(old.User).Pop(v.ID)
		if !found {
			panic(fmt.Errorf("UserVisit of user %d for visit %d has been lost",
				old.User, v.ID))
		}
		db.GetUserVisits(v.User).Add(visit)
	}

	// move mark to new location
	if old.Location != v.Location {
		mark, found := db.GetLocationMarks(old.Location).Pop(v.ID)
		if !found {
			panic(fmt.Errorf("LocationMark of location %d for visit %d has been lost",
				old.Location, v.ID))
		}
		db.GetLocationMarks(v.Location).Add(mark)
	}

	lm := db.GetLocationMarks(v.Location)
	lm.M.Lock()
	for i := range lm.Marks {
		if lm.Marks[i].Visit == v.ID {
			lm.Marks[i].User = v.User
			lm.Marks[i].VisitedAt = v.VisitedAt
			lm.Marks[i].Mark = v.Mark
			break
		}
	}
	lm.M.Unlock()

	uv := db.GetUserVisits(v.User)
	uv.M.Lock()
	for i := range uv.Visits {
		if uv.Visits[i].Visit == v.ID {
			uv.Visits[i].Mark = v.Mark
			uv.Visits[i].VisitedAt = v.VisitedAt
			uv.Visits[i].Location = v.Location
			break
		}
	}
	sort.Sort(models.UserVisitByVisitedAt(uv.Visits))
	uv.M.Unlock()

	db.visits.Set(v.ID, v)
}
