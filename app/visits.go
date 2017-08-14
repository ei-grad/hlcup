package app

import (
	"fmt"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/ei-grad/hlcup/models"
)

type UserVisitFilter func(models.UserVisit) bool

// GetVisitsFilter validates query args and returns a function to filter
// UserVisit's based on this parameters
//
// Parameters:
//
//     fromDate - посещения с visited_at > fromDate
//     toDate - посещения с visited_at < toDate
//     country - название страны, в которой находятся интересующие
//               достопримечательности
//     toDistance - возвращать только те места, у которых
//                  расстояние от города меньше этого параметра
//
func GetVisitsFilter(args *fasthttp.Args) (ret UserVisitFilter, err error) {

	var filters []UserVisitFilter

	fromDateRaw := args.Peek("fromDate")
	if fromDateRaw != nil {
		fromDate, err := strconv.Atoi(string(fromDateRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid fromDate: %s", err)
		}
		filters = append(filters, filterUserVisitFromDate(fromDate))
	}

	toDateRaw := args.Peek("toDate")
	if toDateRaw != nil {
		toDate, err := strconv.Atoi(string(toDateRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid toDate: %s", err)
		}
		filters = append(filters, filterUserVisitToDate(toDate))
	}

	countryRaw := args.Peek("country")
	if countryRaw != nil {
		filters = append(filters, filterUserVisitCountry(string(countryRaw)))
	}

	toDistanceRaw := args.Peek("toDistance")
	if toDistanceRaw != nil {
		toDistance, err := strconv.ParseUint(string(toDistanceRaw), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid toDistance: %s", err)
		}
		filters = append(filters, filterUserVisitToDistance(uint32(toDistance)))
	}

	ret = func(v models.UserVisit) bool {
		for _, i := range filters {
			if !i(v) {
				return false
			}
		}
		return true
	}

	return ret, nil
}

func filterUserVisitFromDate(t int) UserVisitFilter {
	return func(v models.UserVisit) bool {
		return v.VisitedAt > t
	}
}

func filterUserVisitToDate(t int) UserVisitFilter {
	return func(v models.UserVisit) bool {
		return v.VisitedAt < t
	}
}

func filterUserVisitCountry(country string) UserVisitFilter {
	return func(v models.UserVisit) bool {
		return v.Country == country
	}
}

func filterUserVisitToDistance(t uint32) UserVisitFilter {
	return func(v models.UserVisit) bool {
		return v.Distance < t
	}
}
