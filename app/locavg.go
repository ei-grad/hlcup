package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/models"
)

type LocationMarkFilter func(models.LocationMark) bool

// GetMarksFilter validates query args and returns a function to filter
// LocationMark's based on this parameters
//
// Parameters:
//
//     fromDate - учитывать оценки только с visited_at > fromDate
//     toDate - учитывать оценки только с visited_at < toDate
//     fromAge - учитывать только путешественников, у которых
//               возраст (считается от текущего timestamp) больше
//               этого параметра
//     toAge - как предыдущее, но наоборот
//     gender - учитывать оценки только мужчин или женщин
//
func GetMarksFilter(args *fasthttp.Args) (ret LocationMarkFilter, err error) {

	var filters []LocationMarkFilter

	fromDateRaw := args.Peek("fromDate")
	if fromDateRaw != nil {
		fromDate, err := strconv.Atoi(string(fromDateRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid fromDate: %s", err)
		}
		filters = append(filters, filterLocationMarkFromDate(fromDate))
	}

	toDateRaw := args.Peek("toDate")
	if toDateRaw != nil {
		toDate, err := strconv.Atoi(string(toDateRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid toDate: %s", err)
		}
		filters = append(filters, filterLocationMarkToDate(toDate))
	}

	fromAgeRaw := args.Peek("fromAge")
	if fromAgeRaw != nil {
		fromAge, err := strconv.ParseUint(string(fromAgeRaw), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid fromAge: %s", err)
		}
		t := time.Now()
		t = time.Date(t.Year()-int(fromAge), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		filters = append(filters, filterLocationMarkFromAge(t))
	}

	toAgeRaw := args.Peek("toAge")
	if toAgeRaw != nil {
		toAge, err := strconv.ParseUint(string(toAgeRaw), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid toAge: %s", err)
		}
		t := time.Now()
		t = time.Date(t.Year()-int(toAge)-1, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		filters = append(filters, filterLocationMarkToAge(t))
	}

	genderRaw := args.Peek("gender")
	if genderRaw != nil {
		if len(genderRaw) != 1 || (genderRaw[0] != 'm' && genderRaw[0] != 'f') {
			return nil, fmt.Errorf("invalid gender")
		}
		filters = append(filters, filterLocationMarkCountry(genderRaw[0]))
	}

	ret = func(v models.LocationMark) bool {
		for _, i := range filters {
			if !i(v) {
				return false
			}
		}
		return true
	}

	return ret, nil
}

func filterLocationMarkFromDate(t int) LocationMarkFilter {
	return func(v models.LocationMark) bool {
		return v.VisitedAt > t
	}
}

func filterLocationMarkToDate(t int) LocationMarkFilter {
	return func(v models.LocationMark) bool {
		return v.VisitedAt < t
	}
}

func filterLocationMarkFromAge(t time.Time) LocationMarkFilter {
	return func(v models.LocationMark) bool {
		return v.BirthDate.Before(t)
	}
}

func filterLocationMarkToAge(t time.Time) LocationMarkFilter {
	return func(v models.LocationMark) bool {
		return !v.BirthDate.Before(t)
	}
}

func filterLocationMarkCountry(gender byte) LocationMarkFilter {
	return func(v models.LocationMark) bool {
		return v.Gender == gender
	}
}
