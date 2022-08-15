package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type INotificationRepository interface {
	Create(*models.Notification) (*primitive.ObjectID, error)
	GetByCompany(primitive.ObjectID) (*responses.Notification, error)
}

type notificationRepository struct {
}

func NewNotificationRepository() INotificationRepository {
	return &notificationRepository{}
}

var notificationCollection = database.GetInstance().Collection("notifications")

// Create a new tet case
func (r *notificationRepository) Create(notification *models.Notification) (*primitive.ObjectID, error) {

	_, err := notificationCollection.DeleteMany(context.TODO(),
		queries.DeleteNotificationbyCompany(notification.Company))
	if err != nil {
		return nil, err
	}
	res, err := notificationCollection.InsertOne(context.TODO(), notification)
	if err != nil {
		return nil, err
	}

	id := res.InsertedID.(primitive.ObjectID)
	return &id, nil
}

// Get returns a list of all test cases
func (r *notificationRepository) GetByCompany(companyId primitive.ObjectID) (*responses.Notification, error) {

	var notification *responses.Notification
	err := notificationCollection.FindOne(context.TODO(),
		queries.GetNotifictionByCompany(companyId)).Decode(&notification)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return notification, nil
}
