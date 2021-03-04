package models

import (
	"gopkg.in/mgo.v2/bson"
)

// User model
type User struct {
	ID         bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string        `json:"name" bson:"name" binding:"required"`
	Email      string        `json:"email" bson:"email" binding:"required" vañidate:"required,email"`
	Password   string        `json:"password" bson:"password" binding:"required"`
	Phone      string        `json:"phone" bson:"phone"`
	IsVerified bool          `json:"is_verified" bson:"is_verified"`
}
