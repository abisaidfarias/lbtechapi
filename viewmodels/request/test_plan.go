package request

import "go.mongodb.org/mongo-driver/bson/primitive"

type TestPlan struct {
	Name           string             `json:"name" binding:"required"`
	TestCategories []string           `json:"test_categories" binding:"required"`
	Description    string             `json:"description"`
	UserID         primitive.ObjectID `bson:"user_id"`
}
