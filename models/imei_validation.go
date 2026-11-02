package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ImeiQuerySourceLocal    = "LOCAL"
	ImeiQuerySourceImeiInfo = "IMEI_INFO"
	ImeiQuerySourceError    = "ERROR"
)

// TacCatalog caches device identity data keyed by TAC (first 8 IMEI digits).
type TacCatalog struct {
	TAC             string             `bson:"_id"`
	Brand           string             `bson:"brand"`
	TechnicalModel  string             `bson:"technical_model"`
	CommercialName  string             `bson:"commercial_name"`
	SourceIMEI      string             `bson:"source_imei"`
	RawResponse     primitive.M        `bson:"raw_response,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"`
}

// ImeiQuery is one lookup attempt stored for the Historial screen.
type ImeiQuery struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	IMEI            string             `bson:"imei"`
	TAC             string             `bson:"tac"`
	Brand           string             `bson:"brand,omitempty"`
	TechnicalModel  string             `bson:"technical_model,omitempty"`
	CommercialName  string             `bson:"commercial_name,omitempty"`
	Source          string             `bson:"source"`
	ErrorMessage    string             `bson:"error_message,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id"`
	QueriedAt       time.Time          `bson:"queried_at"`
}
