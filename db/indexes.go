package db

import (
	"fmt"

	"github.com/ei-grad/hlcup/models"
)

func (db *DB) GetLocationMarks(id uint32) *models.LocationMarks {
	lm := db.locationMarks.Get(id)
	if lm == nil {
		t, _ := db.sf.Do(fmt.Sprintf("locationMarks/%d", id), func() (interface{}, error) {
			// check for race condition between first Get and singleflight run
			ret := db.locationMarks.Get(id)
			if ret != nil {
				// the other singleflight has already run, just return its result
				return ret, nil
			}
			// ok, it is the first singleflight run
			ret = &models.LocationMarks{}
			db.locationMarks.Set(id, ret)
			return ret, nil

		})
		lm, _ = t.(*models.LocationMarks)
	}
	return lm
}

func (db *DB) GetUserVisits(id uint32) *models.UserVisits {
	uv := db.userVisits.Get(id)
	if uv == nil {
		t, _ := db.sf.Do(fmt.Sprintf("userVisits/%d", id), func() (interface{}, error) {
			// check for race condition between first Get and singleflight run
			ret := db.userVisits.Get(id)
			if ret != nil {
				// the other singleflight has already run, just return its result
				return ret, nil
			}
			// ok, it is the first singleflight run
			ret = &models.UserVisits{}
			db.userVisits.Set(id, ret)
			return ret, nil
		})
		uv, _ = t.(*models.UserVisits)
	}
	return uv
}
