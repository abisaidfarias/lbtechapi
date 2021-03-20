package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TestPlan struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `json:"name"  `
	TotalCategory int32              `json:"total_category"`
	TotalTest     int32              `json:"total_test"`
	Description   string             `json:"description"`
	User          string             `json:"user"`
}
