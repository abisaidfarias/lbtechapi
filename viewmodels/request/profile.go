package request

// Profile model
type Profile struct {
	Name       string  `bson:"name,omitempty" binding:"required"`
	IsInternal bool    `bson:"is_internal" json:"is_internal"`
	Claims     []Claim `bson:"claims" binding:"required"`
}
