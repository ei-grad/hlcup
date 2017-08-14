package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/models"
)

func (app *Application) getEntity(ctx *fasthttp.RequestCtx, entity string, id uint32) int {

	var v interface {
		IsValid() bool
		MarshalJSON() ([]byte, error)
	}

	switch entity {
	case strUsers:
		user := app.db.GetUser(id)
		v = &user
	case strLocations:
		location := app.db.GetLocation(id)
		v = &location
	case strVisits:
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

	ctx.Write(resp)

	return http.StatusOK

}

func (app *Application) getUserVisits(ctx *fasthttp.RequestCtx, id uint32) int {

	if !app.db.GetUser(id).IsValid() {
		return http.StatusNotFound
	}

	filter, err := GetVisitsFilter(ctx.QueryArgs())
	if err != nil {
		return http.StatusBadRequest
	}

	first := true

	ctx.WriteString(`{"visits":[`)

	visits := app.db.GetUserVisits(id)
	visits.M.RLock()
	for _, i := range visits.Visits {
		// TODO: implement /users/<id>/visits filters
		if !filter(i) {
			continue
		}
		if !first {
			ctx.WriteString(",")
		}
		tmp, _ := i.MarshalJSON()
		ctx.Write(tmp)
		first = false
	}
	visits.M.RUnlock()

	ctx.WriteString("]}")

	return http.StatusOK
}

func (app *Application) getLocationAvg(ctx *fasthttp.RequestCtx, id uint32) int {

	if !app.db.GetLocation(id).IsValid() {
		return http.StatusNotFound
	}

	filter, err := GetMarksFilter(ctx.QueryArgs())
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

	ctx.WriteString(fmt.Sprintf(`{"avg": %.5f}`, avg))

	return http.StatusOK
}

func (app *Application) getLocationMarks(ctx *fasthttp.RequestCtx, id uint32) int {

	if !app.db.GetLocation(id).IsValid() {
		return http.StatusNotFound
	}

	first := true

	ctx.WriteString(`{"marks":[`)

	marks := app.db.GetLocationMarks(id)
	marks.M.RLock()
	for _, i := range marks.Marks {
		if !first {
			ctx.WriteString(",")
		}
		tmp, _ := i.MarshalJSON()
		ctx.Write(tmp)
		first = false
	}
	marks.M.RUnlock()

	ctx.WriteString("]}")

	return http.StatusOK
}

func (app *Application) postEntityNew(ctx *fasthttp.RequestCtx, entity string, body []byte) int {

	var v interface {
		UnmarshalJSON([]byte) error
		Validate() error
	}

	var saver func() error

	switch entity {
	case strUsers:
		var user models.User
		v = &user
		saver = func() error { return app.db.AddUser(user) }
	case strLocations:
		var location models.Location
		v = &location
		saver = func() error { return app.db.AddLocation(location) }
	case strVisits:
		var visit models.Visit
		v = &visit
		saver = func() error { return app.db.AddVisit(visit) }
	default:
		return http.StatusNotFound
	}

	if err := v.UnmarshalJSON(body); err != nil {
		ctx.Logger().Printf(err.Error())
		return http.StatusBadRequest
	}
	if err := v.Validate(); err != nil {
		ctx.Logger().Printf("validate failed: %s", err.Error())
		return http.StatusBadRequest
	}
	if err := saver(); err != nil {
		ctx.Logger().Printf("can't add %+v: %s\nBody:\n%s", v, err.Error(), body)
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func (app *Application) postEntity(ctx *fasthttp.RequestCtx, entity string, id uint32, body []byte) int {

	var (
		v interface {
			Validate() error
			IsValid() bool
		}
		user     models.User
		location models.Location
		visit    models.Visit
		err      error
	)

	switch entity {
	case strUsers:
		user = app.db.GetUser(id)
		v = &user
	case strLocations:
		location = app.db.GetLocation(id)
		v = &location
	case strVisits:
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
	case strUsers:
		err = user.UnmarshalJSON(body)
		if err == nil && user.ID != id {
			err = errors.New("id is forbidden in update")
		}
	case strLocations:
		err = location.UnmarshalJSON(body)
		if err == nil && location.ID != id {
			err = errors.New("id is forbidden in update")
		}
	case strVisits:
		err = visit.UnmarshalJSON(body)
		if err == nil && visit.ID != id {
			err = errors.New("id is forbidden in update")
		}
	}

	if err == nil {
		err = v.Validate()
	}

	if err != nil {
		return http.StatusBadRequest
	}

	switch entity {
	case strUsers:
		app.db.UpdateUser(user)
	case strLocations:
		app.db.UpdateLocation(location)
	case strVisits:
		app.db.UpdateVisit(visit)
	}

	return http.StatusOK
}
