package db

import (
	"github.com/ei-grad/hlcup/models"
)

func (db *DB) GetLocationMarks(id uint32) *models.LocationMarks {
	db.lockLM.RLock(id)
	lm := db.locationMarks[id]
	db.lockLM.RUnlock(id)
	if lm == nil {
		db.lockLM.Lock(id)
		// check for the race condition
		lm = db.locationMarks[id]
		if lm == nil {
			lm = &models.LocationMarks{}
			db.locationMarks[id] = lm
		}
		db.lockLM.Unlock(id)
	}
	return lm
}

func (db *DB) GetUserVisits(id uint32) *models.UserVisits {
	db.lockUV.RLock(id)
	uv := db.userVisits[id]
	db.lockUV.RUnlock(id)
	if uv == nil {
		db.lockUV.Lock(id)
		// check for the race condition
		uv = db.userVisits[id]
		if uv == nil {
			uv = &models.UserVisits{}
			db.userVisits[id] = uv
		}
		db.lockUV.Unlock(id)
	}
	return uv
}
