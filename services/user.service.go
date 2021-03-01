package services

import(
	"github.com/abisaidfarias/lbtechapi/models"
)

// UserService interface 
type UserService interface{
	Save(models.User) models.User
	FindAll() []models.User
}

type userService struct{
	users []models.User
}
// New implement
func New() UserService{
	return &userService{}
}

func (services *userService) Save(user models.User) models.User{
	services.users = append(services.users,user)
	return user
}

func (services *userService) FindAll() []models.User{
	return services.users
}

