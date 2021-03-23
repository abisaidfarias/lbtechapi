package request

// Profile model
type Profile struct {
	Name       string  `bson:"name,omitempty" binding:"required"`
	IsInternal bool    `json:"is_internal" binding:"required"`
	Claims     []Claim `bson:"claims" binding:"required"`
}
