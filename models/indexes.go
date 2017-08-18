package models

import (
	"sort"
	"sync"
	"time"
)

//go:generate easyjson $GOFILE

// LocationMark contains info needed to implement filters:
//    fromDate - учитывать оценки только с visited_at > fromDate
//    toDate - учитывать оценки только с visited_at < toDate
//    fromAge - учитывать только путешественников, у которых возраст (считается от текущего timestamp) больше этого параметра
//    toAge - как предыдущее, но наоборот
//    gender - учитывать оценки только мужчин или женщин
//easyjson:json
type LocationMark struct {
	Visit     uint32
	User      uint32
	VisitedAt int
	BirthDate time.Time
	Gender    byte
	Mark      uint8
}

// LocationMarks is used to calculate average location mark
type LocationMarks struct {
	M     sync.RWMutex
	Marks []LocationMark
}

func (lm *LocationMarks) Add(m LocationMark) {
	lm.M.Lock()
	lm.Marks = append(lm.Marks, m)
	lm.M.Unlock()
}

func (lm *LocationMarks) Pop(visitID uint32) (LocationMark, bool) {
	lm.M.Lock()
	defer lm.M.Unlock()
	for n, i := range lm.Marks {
		if i.Visit == visitID {
			lm.Marks = append(lm.Marks[:n], lm.Marks[n+1:]...)
			return i, true
		}
	}
	return LocationMark{}, false
}

// UserVisit is used to filter and output the user visit info
//    fromDate - посещения с visited_at > fromDate
//    toDate - посещения с visited_at < toDate
//    country - название страны, в которой находятся интересующие достопримечательности
//    toDistance - возвращать только те места, у которых расстояние от города меньше этого параметра
//easyjson:json
type UserVisit struct {
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`
	Country   string `json:"-"`
	Visit     uint32 `json:"-"`
	Location  uint32 `json:"-"`
	Distance  uint32 `json:"-"`
	Mark      uint8  `json:"mark"`
}

// UserVisits is user visits index
type UserVisits struct {
	M      sync.RWMutex
	Visits []UserVisit
}

type UserVisitByVisitedAt []UserVisit

// Len is part of sort.Interface.
func (uv UserVisitByVisitedAt) Len() int {
	return len(uv)
}

// Swap is part of sort.Interface.
func (uv UserVisitByVisitedAt) Swap(i, j int) {
	uv[i], uv[j] = uv[j], uv[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (uv UserVisitByVisitedAt) Less(i, j int) bool {
	return uv[i].VisitedAt < uv[j].VisitedAt
}

func (uv *UserVisits) Add(v UserVisit) {
	uv.M.Lock()
	uv.Visits = append(uv.Visits, v)
	sort.Sort(UserVisitByVisitedAt(uv.Visits))
	uv.M.Unlock()
}

func (uv *UserVisits) Pop(visitID uint32) (UserVisit, bool) {
	uv.M.Lock()
	defer uv.M.Unlock()
	for n, i := range uv.Visits {
		if i.Visit == visitID {
			uv.Visits = append(uv.Visits[:n], uv.Visits[n+1:]...)
			return i, true
		}
	}
	return UserVisit{}, false
}
