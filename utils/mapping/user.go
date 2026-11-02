package mapping

import (
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func userRequestBase(userRequest *request.User) *models.User {
	brands := functions.StringsToObjectIds(userRequest.Brands)
	countries := functions.StringsToObjectIds(userRequest.Countries)
	clients := functions.StringsToObjectIds(userRequest.Clients)

	var company primitive.ObjectID
	var profile primitive.ObjectID

	if userRequest.Company != "" {
		company, _ = primitive.ObjectIDFromHex(userRequest.Company)
	}
	if userRequest.Profile != "" {
		profile, _ = primitive.ObjectIDFromHex(userRequest.Profile)
	}
	return &models.User{
		Email:      userRequest.Email,
		Name:       userRequest.Name,
		LastName:   userRequest.LastName,
		Phone:      userRequest.Phone,
		IsInternal: userRequest.IsInternal,
		Brands:     brands,
		Countries:  countries,
		Company:    company,
		Profile:    profile,
		Clients:    clients,
	}
}

func UserRequestToPendingUser(userRequest *request.User) *models.User {
	user := userRequestBase(userRequest)
	user.AccountStatus = models.AccountStatusPending
	user.PasswordHash = ""
	return user
}

func UserRequestToActiveUser(userRequest *request.User) *models.User {
	user := userRequestBase(userRequest)
	user.AccountStatus = models.AccountStatusActive
	user.PasswordHash = functions.HashPassword(userRequest.Password)
	now := time.Now()
	user.PasswordSetAt = &now
	return user
}

// UserRequestToUser keeps upgrade/bootstrap behavior with explicit password.
func UserRequestToUser(userRequest *request.User) *models.User {
	if strings.TrimSpace(userRequest.Password) == "" {
		return UserRequestToPendingUser(userRequest)
	}
	return UserRequestToActiveUser(userRequest)
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
