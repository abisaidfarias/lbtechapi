package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Company struct {
	ID      primitive.ObjectID `bson:"_id" json:"_id"`
	Email   string             `bson:"email" json:"email"`
	Name    string             `bson:"name" json:"name"`
	Address string             `bson:"address" json:"address"`
	LogoUrl string             `json:"logo_url" bson:"logo_url"`
}
