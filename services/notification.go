package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// INotificationService is the notification service
type INotificationService interface {
	Create(*request.Notification) (string, error)
	GetByCompany(string) (*responses.Notification, error)
}

type notificationService struct {
	notificationRepository repositories.INotificationRepository
}

// NewNotificationService is a constructor
func NewNotificationService(notificationRepository repositories.INotificationRepository) INotificationService {
	return &notificationService{
		notificationRepository: notificationRepository,
	}
}

// Create creates a new cateogry
func (s *notificationService) Create(notificationRequest *request.Notification) (string, error) {

	notification := mapping.NotificationRequestToNotification(notificationRequest)

	id, err := s.notificationRepository.Create(notification)

	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

// Get gets a list of all categories
func (s *notificationService) GetByCompany(id string) (*responses.Notification, error) {
	companyId, _ := primitive.ObjectIDFromHex(id)

	result, err := s.notificationRepository.GetByCompany(companyId)

	if err != nil {
		return nil, err
	}

	return result, nil
}
