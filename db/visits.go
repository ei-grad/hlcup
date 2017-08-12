package db

import (
	"strconv"

	"github.com/ei-grad/hlcup/models"
)

// AddVisit adds Visit to index
func (db *DB) AddVisit(v models.Visit) {

	location := db.GetLocation(v.Location)
	//user := db.GetUser(v.User)

	uv := db.userVisits.Get(v.User)
	if uv == nil {
		t, _ := db.userSF.Do(strconv.Itoa(int(v.User)), func() (interface{}, error) {
			ret := &models.UserVisits{
				Visits: []models.UserVisit{},
			}
			db.userVisits.Set(v.User, ret)
			return ret, nil
		})
		uv, _ = t.(*models.UserVisits)
	}
	uv.Visits = append(uv.Visits, models.UserVisit{
		Mark:      v.Mark,
		VisitedAt: v.VisitedAt,
		Place:     location.City,
		VisitID:   v.ID,
		Country:   location.Country,
		Distance:  location.Distance,
	})

}
