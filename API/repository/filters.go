package repository

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDateFilters(req bson.M, startDate string, endDate string) error {
	if len(startDate) > 0 || len(endDate) > 0 {
		lastUpdateReq := bson.M{}
		if len(startDate) > 0 {
			startDate, e := time.Parse("2006-01-02", startDate)
			if e != nil {
				return e
			}
			lastUpdateReq["$gte"] = primitive.NewDateTimeFromTime(startDate)
		}

		if len(endDate) > 0 {
			parsedEndDate, e := time.Parse("2006-01-02", endDate)
			parsedEndDate = parsedEndDate.Add(time.Hour * 24)
			if e != nil {
				return e
			}
			lastUpdateReq["$lte"] = primitive.NewDateTimeFromTime(parsedEndDate)
		}
		req["lastUpdated"] = lastUpdateReq
	}
	return nil
}
