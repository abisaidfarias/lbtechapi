package queries

import "go.mongodb.org/mongo-driver/bson"

// SortByOngoingStatusThenPlanningDateDesc orders list results with ongoing records
// first (status ascending), then by planning_date from newest to oldest.
func SortByOngoingStatusThenPlanningDateDesc() bson.D {
	return bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "status", Value: 1},
			{Key: "planning_date", Value: -1},
			{Key: "created_date", Value: -1},
		}},
	}
}
