package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TestPlan struct {
	ID             primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Name           string               `json:"name"  `
	TotalCategory  int                  `json:"total_category"`
	TotalTest      int                  `json:"total_test"`
	Description    string               `json:"description"`
	UserName       string               `json:"userName"`
	TestCategories []primitive.ObjectID `bson:"test_categories,omitempty" json:"test_categories"`
}
