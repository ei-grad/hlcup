package db

import (
	"fmt"
	"sort"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) UpdateUser(v models.User) {

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

	db.users.Set(v.ID, v)
}

func (db *DB) UpdateLocation(v models.Location) {

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
			for i := range uv.Visits {
				if uv.Visits[i].Location == v.ID {
					uv.Visits[i].Place = v.Place
					uv.Visits[i].Country = v.Country
					uv.Visits[i].Distance = v.Distance
				}
			}
			uv.M.Unlock()
		}
	}

	db.locations.Set(v.ID, v)
}

func (db *DB) UpdateVisit(v models.Visit) {

	old := db.GetVisit(v.ID)

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
			uv.Visits[n] = models.UserVisit{
				Visit:     v.ID,
				Location:  v.Location,
				Mark:      v.Mark,
				VisitedAt: v.VisitedAt,
				Place:     location.Place,
				Country:   location.Country,
				Distance:  location.Distance,
			}
			break
		}
	}
	sort.Sort(models.UserVisitByVisitedAt(uv.Visits))
	uv.M.Unlock()

	db.visits.Set(v.ID, v)
}
