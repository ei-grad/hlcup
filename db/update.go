package db

import (
	"log"
	"sort"
	"time"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) UpdateUser(v models.User) {

	old := db.users.Get(v.ID)

	if old.BirthDate != v.BirthDate || old.Gender != v.Gender {
		ul := db.userLocations.Get(v.ID)
		if ul != nil {
			ul.M.RLock()
			for _, i := range ul.Locations {
				lm := db.locationMarks.Get(i)
				if lm == nil {
					continue
				}
				lm.M.Lock()
				for _, i := range lm.Marks {
					if i.User == v.ID {
						i.BirthDate = time.Unix(v.BirthDate, 0)
						i.Gender = []byte(v.Gender)[0]
					}
				}
				lm.M.Unlock()
			}
			ul.M.RUnlock()
		}
	}

	db.users.Set(v.ID, v)
}

func (db *DB) UpdateLocation(v models.Location) {

	old := db.locations.Get(v.ID)

	if old.Place != v.Place || old.Country != v.Country || old.Distance != v.Distance {
		lu := db.locationUsers.Get(v.ID)
		lu.M.RLock()
		for _, i := range lu.Users {
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
		lu.M.RUnlock()
	}

	db.locations.Set(v.ID, v)
}

func (db *DB) UpdateVisit(v models.Visit) {

	old := db.visits.Get(v.ID)

	// TODO: fuck that shit!
	if old.User != v.User {
		log.Printf("visit %d changed user %d -> %d", v.ID, old.User, v.User)
	}

	// TODO: fuck that shit!
	if old.Location != v.Location {
		log.Printf("visit %d changed location %d -> %d", v.ID, old.Location, v.Location)
	}

	if old.Mark != v.Mark {

		lm := db.locationMarks.Get(v.Location)
		lm.M.Lock()
		for _, i := range lm.Marks {
			if i.Visit == v.ID {
				i.Mark = v.Mark
			}
		}
		lm.M.Unlock()

		uv := db.userVisits.Get(v.User)
		uv.M.Lock()
		for _, i := range uv.Visits {
			if i.Visit == v.ID {
				i.Mark = v.Mark
				i.VisitedAt = v.VisitedAt
			}
		}
		sort.Sort(UserVisitByVisitedAt(uv.Visits))
		uv.M.Unlock()

	}

	db.visits.Set(v.ID, v)
}
