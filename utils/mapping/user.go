package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UserRequestToUser(userRequest *request.User) *models.User {

	hashedPassword := functions.HashPassword(userRequest.Password)
	brands := functions.StringsToObjectIds(userRequest.Brands)
	countries := functions.StringsToObjectIds(userRequest.Countries)
	clients := functions.StringsToObjectIds(userRequest.Clients)
	company, _ := primitive.ObjectIDFromHex(userRequest.Company)
	profile, _ := primitive.ObjectIDFromHex(userRequest.Profile)
	return &models.User{
		Email:        userRequest.Email,
		PasswordHash: hashedPassword,
		Name:         userRequest.Name,
		LastName:     userRequest.LastName,
		Phone:        userRequest.Phone,
		IsInternal:   userRequest.IsInternal,
		Brands:       brands,
		Countries:    countries,
		Company:      company,
		Profile:      profile,
		Clients:      clients,
	}
}

func UserToUserResponse(user *models.User) *responses.User {

	return &responses.User{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Company:    user.Company,
		IsInternal: user.IsInternal,
		Brands:     user.Brands,
		Countries:  user.Countries,
		Clients:    user.Clients,
	}
}
func UserResumeToUser(userResume *request.UserResume) *models.User {

	id, _ := primitive.ObjectIDFromHex(userResume.UserID)
	return &models.User{
		ID:         id,
		Email:      userResume.Email,
		Name:       userResume.Name,
		LastName:   userResume.LastName,
		IsInternal: userResume.IsInternal,
	}
}
