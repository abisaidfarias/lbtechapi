package request

type ChangePassword struct {
	NewPassword string `json:"new_password" bson:"new_password" binding:"required,passwordFormat"`
	OldPassword string `json:"old_password" bson:"old_password" binding:"required,passwordFormat"`
}
