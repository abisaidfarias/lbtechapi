package services

import (
	"fmt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// IUserService auth interface
type IUserService interface {
	Create(*request.User) error
	GetByID(string) (*responses.User, error)
	Get(string) ([]*bson.M, error)
	GetByEmail(string) (*responses.AuthUser, error)
	GetProfileByID(string) (*responses.Profile, error)
	Update(string, *models.User) error
	Delete(string) error
	ChangePassword(string, request.ChangePassword) error
	Upgrade(*request.User) error
	GetInternalUser() ([]*responses.User, error)
	GetUsersByCompany(string) ([]*responses.User, error)
}
type userService struct {
	userRepository    repositories.IUserRepository
	profileRepository repositories.IProfileRepository
	companyRepository repositories.ICompanyRepository
}

// NewUserService is a constructor
func NewUserService(userRepository repositories.IUserRepository, profileRepository repositories.IProfileRepository, companyRepository repositories.ICompanyRepository) IUserService {
	return &userService{
		userRepository:    userRepository,
		profileRepository: profileRepository,
		companyRepository: companyRepository,
	}
}

// Create creates and saves the new user
func (s *userService) Create(userRequest *request.User) error {

	userLogged, err := s.userRepository.GetByID(userRequest.UserID)

	if err != nil {
		return err
	}
	if userRequest.IsInternal {
		userRequest.Company = userLogged.Company.Hex()
	}

	user := mapping.UserRequestToUser(userRequest)

	err = s.userRepository.Create(user)

	if err != nil {
		return err
	}

	return nil
}
func (s *userService) Upgrade(userRequest *request.User) error {
	// Verificar si ya existen usuarios en el sistema
	fmt.Printf("DEBUG: Verificando conteo de usuarios...\n")
	userCount, err := s.userRepository.Count()
	if err != nil {
		fmt.Printf("DEBUG: Error al contar usuarios: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: Usuarios encontrados: %d\n", userCount)

	// Si ya hay usuarios, no permitir usar el endpoint upgrade
	if userCount > 0 {
		return fmt.Errorf("%w", utils.ErrorUpgradeNotAllowed)
	}

	var profileID primitive.ObjectID
	var companyID primitive.ObjectID

	// Si se proporciona un Profile, validar que exista
	if userRequest.Profile != "" {
		profile, err := s.profileRepository.GetById(userRequest.Profile)
		if err != nil {
			return fmt.Errorf("profile with id '%s' does not exist: %w", userRequest.Profile, utils.ErrorResourceNotFound)
		}
		profileID = profile.ID
	} else {
		// Si no se proporciona Profile, verificar si hay perfiles
		profileCount, err := s.profileRepository.Count()
		if err != nil {
			return err
		}

		if profileCount == 0 {
			// Crear perfil por defecto
			fmt.Printf("DEBUG: Creando perfil por defecto...\n")
			defaultProfile := &models.Profile{
				Name:       "Admin",
				IsInternal: true,
				Claims:     []models.Claim{}, // Perfil sin claims específicos por defecto
			}
			err = s.profileRepository.Create(defaultProfile)
			if err != nil {
				fmt.Printf("DEBUG: Error al crear perfil: %v\n", err)
				return fmt.Errorf("failed to create default profile: %w", err)
			}
			fmt.Printf("DEBUG: Perfil creado con ID: %s\n", defaultProfile.ID.Hex())
			profileID = defaultProfile.ID
		} else {
			// Si hay perfiles pero no se proporcionó uno, usar el primero disponible
			profiles, err := s.profileRepository.Get()
			if err != nil || len(profiles) == 0 {
				return fmt.Errorf("no profiles available")
			}
			// Obtener el ID del primer perfil desde bson.M
			firstProfile := *profiles[0]
			idValue := firstProfile["_id"]
			if idValue == nil {
				return fmt.Errorf("profile has no _id field")
			}
			// Convertir a ObjectID (puede venir como ObjectID o como string)
			switch v := idValue.(type) {
			case primitive.ObjectID:
				profileID = v
			case string:
				oid, err := primitive.ObjectIDFromHex(v)
				if err != nil {
					return fmt.Errorf("invalid profile id format: %w", err)
				}
				profileID = oid
			default:
				return fmt.Errorf("invalid profile id type")
			}
		}
	}

	// Si se proporciona una Company, validar que exista
	if userRequest.Company != "" {
		company, err := s.companyRepository.GetById(userRequest.Company)
		if err != nil {
			return fmt.Errorf("company with id '%s' does not exist: %w", userRequest.Company, utils.ErrorResourceNotFound)
		}
		companyID = company.ID
	} else {
		// Si no se proporciona Company, verificar si hay compañías
		companyCount, err := s.companyRepository.Count()
		if err != nil {
			return err
		}

		if companyCount == 0 {
			// Crear compañía por defecto usando el email del usuario
			fmt.Printf("DEBUG: Creando compañía por defecto...\n")
			defaultCompany := &models.Company{
				Name:    "Default Company",
				Email:   userRequest.Email,
				Address: "",
				LogoUrl: "",
			}
			err = s.companyRepository.Create(defaultCompany)
			if err != nil {
				fmt.Printf("DEBUG: Error al crear compañía: %v\n", err)
				return fmt.Errorf("failed to create default company: %w", err)
			}
			fmt.Printf("DEBUG: Compañía creada con ID: %s\n", defaultCompany.ID.Hex())
			companyID = defaultCompany.ID
		} else {
			// Si hay compañías pero no se proporcionó una, usar la primera disponible
			companies, err := s.companyRepository.Get()
			if err != nil || len(companies) == 0 {
				return fmt.Errorf("no companies available")
			}
			companyID = companies[0].ID
		}
	}

	// Asignar los IDs obtenidos/creados al request
	fmt.Printf("DEBUG: Asignando IDs - Profile: %s, Company: %s\n", profileID.Hex(), companyID.Hex())
	userRequest.Profile = profileID.Hex()
	userRequest.Company = companyID.Hex()

	fmt.Printf("DEBUG: Mapeando usuario...\n")
	user := mapping.UserRequestToUser(userRequest)

	fmt.Printf("DEBUG: Creando usuario...\n")
	err = s.userRepository.Create(user)

	if err != nil {
		fmt.Printf("DEBUG: Error al crear usuario: %v\n", err)
		return err
	}

	fmt.Printf("DEBUG: Usuario creado exitosamente\n")
	return nil
}
func (s *userService) Get(userID string) ([]*bson.M, error) {

	user, err := s.userRepository.GetByID(userID)

	if err != nil {
		return nil, err
	}
	if user.IsInternal {
		users, err := s.userRepository.Get()
		if err != nil {
			return nil, err
		}
		return users, err
	}
	users, err := s.userRepository.GetByCompany(user.Company)
	if err != nil {
		return nil, err
	}
	return users, err
}

// GetByID gets an user by ID
func (s *userService) GetByID(id string) (*responses.User, error) {

	user, err := s.userRepository.GetByID(id)

	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s *userService) GetByEmail(email string) (*responses.AuthUser, error) {

	user, err := s.userRepository.GetByEmail(email)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(id string, user *models.User) error {

	err := s.userRepository.Update(id, user)

	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Delete(id string) error {

	err := s.userRepository.Delete(id)

	if err != nil {
		return err
	}
	return nil
}
func (s *userService) GetProfileByID(id string) (*responses.Profile, error) {

	oid, _ := primitive.ObjectIDFromHex(id)
	profile, err := s.userRepository.GetProfileByID(oid)

	if err != nil {
		return nil, err
	}
	return profile, nil
}
func (s *userService) ChangePassword(email string, changePassword request.ChangePassword) error {

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("%w", utils.ErrorResourceNotFound)
	}
	err = functions.ValidateUserCredentials(user.PasswordHash, changePassword.OldPassword)
	if err != nil {
		return fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	hashedPassword := functions.HashPassword(changePassword.NewPassword)

	err = s.userRepository.ChangePassword(hashedPassword, email)

	if err != nil {
		return err
	}
	return nil
}
func (s *userService) GetInternalUser() ([]*responses.User, error) {

	users, err := s.userRepository.GetInternalUser()

	if err != nil {
		return nil, err
	}
	return users, nil
}
func (s *userService) GetUsersByCompany(id string) ([]*responses.User, error) {

	company, _ := primitive.ObjectIDFromHex(id)
	users, err := s.userRepository.GetUsersByCompany(company)

	if err != nil {
		return nil, err
	}
	return users, nil
}
