package services

import (
	"fmt"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IPasswordTokenService interface {
	IssueAndEmail(userID primitive.ObjectID, email, name, purpose string) error
	RevokeAllForUser(userID primitive.ObjectID) error
	ValidateRawToken(raw string) (*models.PasswordToken, error)
	Consume(raw, newPassword string) error
}

type passwordTokenService struct {
	tokenRepo repositories.IPasswordTokenRepository
	userRepo  repositories.IUserRepository
}

func NewPasswordTokenService(
	tokenRepo repositories.IPasswordTokenRepository,
	userRepo repositories.IUserRepository,
) IPasswordTokenService {
	return &passwordTokenService{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
	}
}

func (s *passwordTokenService) RevokeAllForUser(userID primitive.ObjectID) error {
	return s.tokenRepo.InvalidateAllForUser(userID)
}

func (s *passwordTokenService) IssueAndEmail(userID primitive.ObjectID, email, name, purpose string) error {
	raw, hash, err := functions.GeneratePasswordToken()
	if err != nil {
		return err
	}
	now := time.Now()
	if err := s.tokenRepo.InvalidateActive(userID, purpose); err != nil {
		return err
	}
	token := &models.PasswordToken{
		UserID:    userID,
		TokenHash: hash,
		Purpose:   purpose,
		ExpiresAt: functions.PasswordTokenExpiresAt(now),
		CreatedAt: now,
	}
	if err := s.tokenRepo.Create(token); err != nil {
		return err
	}
	return functions.SendPasswordActionEmail(email, name, purpose, raw)
}

func (s *passwordTokenService) ValidateRawToken(raw string) (*models.PasswordToken, error) {
	hash := functions.HashPasswordToken(raw)
	token, err := s.tokenRepo.FindValidByHash(hash)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w", utils.ErrorInvalidPasswordToken)
		}
		return nil, err
	}
	return token, nil
}

func (s *passwordTokenService) Consume(raw, newPassword string) error {
	token, err := s.ValidateRawToken(raw)
	if err != nil {
		return err
	}
	hashPassword := functions.HashPassword(newPassword)
	activatedAt := time.Now()
	if err := s.userRepo.ActivatePassword(token.UserID, hashPassword, activatedAt); err != nil {
		return err
	}
	return s.tokenRepo.MarkUsed(token.ID)
}
