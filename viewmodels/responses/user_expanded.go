package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserResponse response to client
type UserExpanded struct {
	ID       primitive.ObjectID `bson:"_id"`
	Email    string             `bson:"email"`
	Name     string             `bson:"name"`
	LastName string             `bson:"lastName"`
	Phone    string             `bson:"phone"`
	Dni      string             `bson:"dni"`
	Company  Company            `bson:"company"`
}
