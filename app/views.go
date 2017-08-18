package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/ei-grad/hlcup/entities"
	"github.com/ei-grad/hlcup/models"
)

func (app *Application) GetEntity(w io.Writer, entity entities.Entity, id uint32) int {

	var v interface {
		IsValid() bool
		MarshalJSON() ([]byte, error)
	}

	switch entity {
	case entities.User:
		user := app.db.GetUser(id)
		v = &user
	case entities.Location:
		location := app.db.GetLocation(id)
		v = &location
	case entities.Visit:
		visit := app.db.GetVisit(id)
		v = &visit
	default:
		return http.StatusNotFound
	}

	if !v.IsValid() {
		return http.StatusNotFound
	}

	resp, err := v.MarshalJSON()
	if err != nil {
		// v.MarshalJSON() failed, shouldn't happen
		panic(err)
	}

	w.Write(resp)

	return http.StatusOK

}

func (app *Application) GetUserVisits(w io.Writer, id uint32, args Peeker) int {

	if !app.db.GetUser(id).IsValid() {
		return http.StatusNotFound
	}

	filter, err := GetVisitsFilter(args)
	if err != nil {
		return http.StatusBadRequest
	}

	first := true

	io.WriteString(w, `{"visits":[`)

	visits := app.db.GetUserVisits(id)
	visits.M.RLock()
	v := visits.Visits
	if filter.fromDateIsSet {
		i := sort.Search(len(v), func(i int) bool { return v[i].VisitedAt > filter.fromDate })
		if i < len(v) {
			v = v[i:]
		} else {
			v = v[:0]
		}
	}
	if filter.toDateIsSet {
		i := sort.Search(len(v), func(i int) bool { return v[i].VisitedAt >= filter.toDate })
		if i < len(v) {
			v = v[:i]
		}
	}
	for _, i := range v {
		if !filter.filter(i) {
			continue
		}
		if !first {
			io.WriteString(w, ",")
		}
		tmp, _ := i.MarshalJSON()
		w.Write(tmp)
		first = false
	}
	visits.M.RUnlock()

	io.WriteString(w, "]}")

	return http.StatusOK
}

func (app *Application) GetLocationAvg(w io.Writer, id uint32, args Peeker) int {

	if !app.db.GetLocation(id).IsValid() {
		return http.StatusNotFound
	}

	filter, err := GetMarksFilter(args)
	if err != nil {
		return http.StatusBadRequest
	}

	var sum, count int
	var avg float64

	marks := app.db.GetLocationMarks(id)
	marks.M.RLock()
	for _, i := range marks.Marks {
		if !filter(i) {
			continue
		}
		sum = sum + int(i.Mark)
		count = count + 1
	}
	marks.M.RUnlock()

	if count == 0 {
		// location have no marks
		avg = 0.
	} else {
		avg = float64(sum) / float64(count)
	}

	io.WriteString(w, fmt.Sprintf(`{"avg": %.5f}`, avg))

	return http.StatusOK
}

func (app *Application) GetLocationMarks(w io.Writer, id uint32) int {

	if !app.db.GetLocation(id).IsValid() {
		return http.StatusNotFound
	}

	first := true

	io.WriteString(w, `{"marks":[`)

	marks := app.db.GetLocationMarks(id)
	marks.M.RLock()
	for _, i := range marks.Marks {
		if !first {
			io.WriteString(w, ",")
		}
		tmp, _ := i.MarshalJSON()
		w.Write(tmp)
		first = false
	}
	marks.M.RUnlock()

	io.WriteString(w, "]}")

	return http.StatusOK
}

func (app *Application) PostEntityNew(entity entities.Entity, body []byte) int {

	var v interface {
		UnmarshalJSON([]byte) error
		Validate() error
		GetID() uint32
	}

	var saver func() error

	switch entity {
	case entities.User:
		var user models.User
		v = &user
		saver = func() error { return app.db.AddUser(user) }
	case entities.Location:
		var location models.Location
		v = &location
		saver = func() error { return app.db.AddLocation(location) }
	case entities.Visit:
		var visit models.Visit
		v = &visit
		saver = func() error { return app.db.AddVisit(visit) }
	default: // entities.Unknown
		return http.StatusNotFound
	}

	if err := v.UnmarshalJSON(body); err != nil {
		return http.StatusBadRequest
	}
	if err := saver(); err != nil {
		return http.StatusBadRequest
	}

	app.cache.Set([]byte(fmt.Sprintf("/%s/%d", entities.GetEntityRoute(entity), v.GetID())), body, 0)

	if app.heat != nil {
		app.heat(entity, v.GetID())
	}

	return http.StatusOK
}

func (app *Application) PostEntity(entity entities.Entity, id uint32, body []byte) int {

	var (
		v interface {
			Validate() error
			IsValid() bool
			GetID() uint32
		}
		user     models.User
		location models.Location
		visit    models.Visit
		err      error
	)

	switch entity {
	case entities.User:
		user = app.db.GetUser(id)
		v = &user
	case entities.Location:
		location = app.db.GetLocation(id)
		v = &location
	case entities.Visit:
		visit = app.db.GetVisit(id)
		v = &visit
	default:
		return http.StatusNotFound
	}

	// check that entity already exist
	if !v.IsValid() {
		return http.StatusNotFound
	}

	switch entity {
	case entities.User:
		err = user.UnmarshalJSON(body)
		if err == nil && user.ID != id {
			err = errors.New("id is forbidden in update")
		}
	case entities.Location:
		err = location.UnmarshalJSON(body)
		if err == nil && location.ID != id {
			err = errors.New("id is forbidden in update")
		}
	case entities.Visit:
		err = visit.UnmarshalJSON(body)
		if err == nil && visit.ID != id {
			err = errors.New("id is forbidden in update")
		}
	}

	if err != nil {
		return http.StatusBadRequest
	}

	switch entity {
	case entities.User:
		err = app.db.UpdateUser(user)
	case entities.Location:
		err = app.db.UpdateLocation(location)
	case entities.Visit:
		err = app.db.UpdateVisit(visit)
	}

	if err != nil {
		return http.StatusBadRequest
	}

	app.cache.Del([]byte(fmt.Sprintf("/%s/%d", entities.GetEntityRoute(entity), v.GetID())))

	if app.heat != nil {
		app.heat(entity, id)
	}

	return http.StatusOK
}
