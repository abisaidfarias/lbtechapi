package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

// Profile model
type DeviceResume struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id"`
	Type                   string             `bson:"type" json:"type"`
	Brand                  Brand             `bson:"brand" json:"brand"`
	CommercialModel        string             `bson:"commercial_model" json:"commercial_model"`
	ImageUrl               string             `bson:"image_url" json:"image_url"`
}
