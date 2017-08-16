package app

import (
	"fmt"
	"strconv"

	"github.com/ei-grad/hlcup/models"
)

type UserVisitFilterData struct {
	fromDateIsSet bool
	fromDate      int
	toDateIsSet   bool
	toDate        int
	filter        UserVisitFilter
}

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
func GetVisitsFilter(args Peeker) (ret UserVisitFilterData, err error) {

	var filters []UserVisitFilter

	fromDateRaw := args.Peek("fromDate")
	if fromDateRaw != nil {
		fromDate, err := strconv.Atoi(string(fromDateRaw))
		if err != nil {
			return ret, fmt.Errorf("invalid fromDate: %s", err)
		}
		ret.fromDateIsSet = true
		ret.fromDate = fromDate
	}

	toDateRaw := args.Peek("toDate")
	if toDateRaw != nil {
		toDate, err := strconv.Atoi(string(toDateRaw))
		if err != nil {
			return ret, fmt.Errorf("invalid toDate: %s", err)
		}
		ret.toDateIsSet = true
		ret.toDate = toDate
	}

	countryRaw := args.Peek("country")
	if countryRaw != nil {
		filters = append(filters, filterUserVisitCountry(string(countryRaw)))
	}

	toDistanceRaw := args.Peek("toDistance")
	if toDistanceRaw != nil {
		toDistance, err := strconv.ParseUint(string(toDistanceRaw), 10, 32)
		if err != nil {
			return ret, fmt.Errorf("invalid toDistance: %s", err)
		}
		filters = append(filters, filterUserVisitToDistance(uint32(toDistance)))
	}

	ret.filter = func(v models.UserVisit) bool {
		for _, i := range filters {
			if !i(v) {
				return false
			}
		}
		return true
	}

	return ret, nil
}

func searchUserVisitFromDate(t int) UserVisitFilter {
	return func(v models.UserVisit) bool {
		return v.VisitedAt > t
	}
}

func searchUserVisitToDate(t int) UserVisitFilter {
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
