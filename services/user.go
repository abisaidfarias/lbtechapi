package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	apputils "github.com/abisaidfarias/lbtechapi/utils"
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
	ChangeMyPassword(string, request.ChangePassword) error
	Upgrade(*request.User) error
	GetInternalUser() ([]*responses.User, error)
	GetUsersByCompany(string) ([]*responses.User, error)
	ResendInvitation(actorUserID, targetUserID string) error
}
type userService struct {
	userRepository       repositories.IUserRepository
	profileRepository    repositories.IProfileRepository
	companyRepository    repositories.ICompanyRepository
	passwordTokenService IPasswordTokenService
}

// NewUserService is a constructor
func NewUserService(
	userRepository repositories.IUserRepository,
	profileRepository repositories.IProfileRepository,
	companyRepository repositories.ICompanyRepository,
	passwordTokenService IPasswordTokenService,
) IUserService {
	return &userService{
		userRepository:       userRepository,
		profileRepository:    profileRepository,
		companyRepository:    companyRepository,
		passwordTokenService: passwordTokenService,
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

	if strings.TrimSpace(userRequest.Password) != "" {
		return fmt.Errorf("%w", apputils.NewValidationError("password must not be sent when creating users; user will receive an email to set password"))
	}

	userRequest.Email = normalizeUserEmail(userRequest.Email)
	taken, err := s.userRepository.IsEmailTaken(userRequest.Email)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%w", utils.ErrorDuplicateUserEmail)
	}

	user := mapping.UserRequestToPendingUser(userRequest)

	err = s.userRepository.Create(user)

	if err != nil {
		if utils.ErrorDuplicatedData(err) {
			return fmt.Errorf("%w", utils.ErrorDuplicateUserEmail)
		}
		return err
	}

	name := strings.TrimSpace(user.Name + " " + user.LastName)
	if err := s.passwordTokenService.IssueAndEmail(user.ID, user.Email, name, models.PasswordTokenPurposeInvite); err != nil {
		log.Printf("invite email failed for user %s: %v", user.Email, err)
		return fmt.Errorf("%w: %v", utils.ErrorUserCreatedInviteFailed, err)
	}

	return nil
}

func (s *userService) ResendInvitation(actorUserID, targetUserID string) error {
	actor, err := s.userRepository.GetByID(actorUserID)
	if err != nil {
		return err
	}
	target, err := s.userRepository.GetByID(targetUserID)
	if err != nil {
		return err
	}
	if !actor.IsInternal && actor.Company != target.Company {
		return fmt.Errorf("%w", utils.ErrorResourceNotFound)
	}
	if !userPendingInvitation(target) {
		return fmt.Errorf("%w", utils.ErrorUserNotPendingInvite)
	}
	rateKey := "invite-resend:" + targetUserID
	if !forgotPasswordLimiter.allow(rateKey, time.Minute) {
		return fmt.Errorf("%w", utils.ErrorInviteResendRateLimited)
	}
	name := strings.TrimSpace(target.Name + " " + target.LastName)
	if err := s.passwordTokenService.IssueAndEmail(target.ID, target.Email, name, models.PasswordTokenPurposeInvite); err != nil {
		return fmt.Errorf("%w: %v", utils.ErrorUserCreatedInviteFailed, err)
	}
	return nil
}

func userPendingInvitation(user *responses.User) bool {
	if user == nil {
		return false
	}
	if user.PasswordSetAt != nil {
		return false
	}
	status := functions.NormalizeListAccountStatus(user.AccountStatus, "")
	return status == models.AccountStatusPending
}

func (s *userService) Upgrade(userRequest *request.User) error {
	// Verificar si ya existen usuarios en el sistema
	userCount, err := s.userRepository.Count()
	if err != nil {
		return err
	}

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
			defaultProfile := &models.Profile{
				Name:       "Admin",
				IsInternal: true,
				Claims:     []models.Claim{}, // Perfil sin claims específicos por defecto
			}
			err = s.profileRepository.Create(defaultProfile)
			if err != nil {
				return fmt.Errorf("failed to create default profile: %w", err)
			}
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
			defaultCompany := &models.Company{
				Name:    "Default Company",
				Email:   userRequest.Email,
				Address: "",
				LogoUrl: "",
			}
			err = s.companyRepository.Create(defaultCompany)
			if err != nil {
				return fmt.Errorf("failed to create default company: %w", err)
			}
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
	userRequest.Profile = profileID.Hex()
	userRequest.Company = companyID.Hex()

	if strings.TrimSpace(userRequest.Password) == "" {
		return fmt.Errorf("%w", apputils.NewValidationError("password is required for initial system upgrade"))
	}

	user := mapping.UserRequestToActiveUser(userRequest)

	userRequest.Email = normalizeUserEmail(userRequest.Email)

	err = s.userRepository.Create(user)

	if err != nil {
		if utils.ErrorDuplicatedData(err) {
			return fmt.Errorf("%w", utils.ErrorDuplicateUserEmail)
		}
		return err
	}

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
	user.Email = normalizeUserEmail(user.Email)

	err := s.userRepository.Update(id, user)

	if err != nil {
		if utils.ErrorDuplicatedData(err) {
			return fmt.Errorf("%w", utils.ErrorDuplicateUserEmail)
		}
		return err
	}
	return nil
}

func normalizeUserEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *userService) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("%w", utils.ErrorResourceNotFound)
	}

	if err := s.userRepository.Delete(id); err != nil {
		return err
	}
	if err := s.passwordTokenService.RevokeAllForUser(oid); err != nil {
		log.Printf("revoke password tokens for deleted user %s: %v", id, err)
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
func (s *userService) ChangeMyPassword(userID string, changePassword request.ChangePassword) error {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}
	return s.changePasswordForEmail(user.Email, changePassword)
}

func (s *userService) ChangePassword(email string, changePassword request.ChangePassword) error {
	return fmt.Errorf("%w", utils.ErrorDeprecatedPasswordEndpoint)
}

func (s *userService) changePasswordForEmail(email string, changePassword request.ChangePassword) error {

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
