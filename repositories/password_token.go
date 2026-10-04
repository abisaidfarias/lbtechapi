package repositories

import (
	"context"
	"time"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IPasswordTokenRepository interface {
	Create(*models.PasswordToken) error
	InvalidateActive(userID primitive.ObjectID, purpose string) error
	FindValidByHash(tokenHash string) (*models.PasswordToken, error)
	MarkUsed(id primitive.ObjectID) error
}

type passwordTokenRepository struct{}

func NewPasswordTokenRepository() IPasswordTokenRepository {
	return &passwordTokenRepository{}
}

var passwordTokenCollection = database.GetInstance().Collection("password_tokens")

func (r *passwordTokenRepository) Create(token *models.PasswordToken) error {
	_, err := passwordTokenCollection.InsertOne(context.TODO(), token)
	return err
}

func (r *passwordTokenRepository) InvalidateActive(userID primitive.ObjectID, purpose string) error {
	filter := queries.InvalidatePasswordTokensFilter(userID, purpose)
	update := queries.InvalidatePasswordTokensUpdate(time.Now())
	_, err := passwordTokenCollection.UpdateMany(context.TODO(), filter, update)
	return err
}

func (r *passwordTokenRepository) FindValidByHash(tokenHash string) (*models.PasswordToken, error) {
	var token models.PasswordToken
	err := passwordTokenCollection.FindOne(context.TODO(), queries.FindValidPasswordTokenFilter(tokenHash)).Decode(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *passwordTokenRepository) MarkUsed(id primitive.ObjectID) error {
	res, err := passwordTokenCollection.UpdateOne(
		context.TODO(),
		queries.MarkPasswordTokenUsedFilter(id),
		queries.MarkPasswordTokenUsedUpdate(time.Now()),
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
