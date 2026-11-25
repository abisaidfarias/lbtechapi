package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IMEI Validator: every uploaded IMEI is stored once (unique imei) and moves
// from "ready" (saved, not yet registered in OABI) to "registered". Uploads
// are grouped in batches that are registered as a whole.
const (
	ImeiValidatorStatusReady      = "ready"
	ImeiValidatorStatusRegistered = "registered"
)

type ImeiBatch struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Code             string             `bson:"code"`
	FileName         string             `bson:"file_name"`
	CreatedBy        primitive.ObjectID `bson:"created_by"`
	CreatedByName    string             `bson:"created_by_name"`
	CreatedAt        time.Time          `bson:"created_at"`
	Total            int                `bson:"total"`
	Status           string             `bson:"status"`
	RegisteredAt     *time.Time         `bson:"registered_at,omitempty"`
	RegisteredBy     primitive.ObjectID `bson:"registered_by,omitempty"`
	RegisteredByName string             `bson:"registered_by_name,omitempty"`
}

type ImeiRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	IMEI         string             `bson:"imei"`
	BatchID      primitive.ObjectID `bson:"batch_id"`
	Status       string             `bson:"status"`
	Serial       string             `bson:"serial"`
	// ImeiCount is the declared number of IMEIs of the device (S/N): 1 single
	// SIM, 2 dual SIM. Needed to rebuild the OABI file.
	ImeiCount    int                `bson:"imei_count"`
	CreatedAt    time.Time          `bson:"created_at"`
	RegisteredAt *time.Time         `bson:"registered_at,omitempty"`
	RegisteredBy primitive.ObjectID `bson:"registered_by,omitempty"`
}
