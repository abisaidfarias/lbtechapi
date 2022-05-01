package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NotificationRequestToNotification(notification *request.Notification) *models.Notification {

	companyId, _ := primitive.ObjectIDFromHex(notification.Company)

	var notificationEmails []models.NotificationEmail
	for _, notificationEmail := range notification.NotificationEmails {

		var newNotificaionEmail models.NotificationEmail
		newNotificaionEmail.Email = notificationEmail.Email
		newNotificaionEmail.Type = notificationEmail.Type
		notificationEmails = append(notificationEmails, newNotificaionEmail)
	}

	return &models.Notification{
		Company:            companyId,
		NotificationEmails: notificationEmails,		
	}
}
