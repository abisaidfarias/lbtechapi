package utils

import (
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestIsDuplicateUserEmailConflict_wrappedSentinel(t *testing.T) {
	err := fmt.Errorf("create user: %w", ErrorDuplicateUserEmail)
	if !IsDuplicateUserEmailConflict(err) {
		t.Fatal("expected duplicate email conflict")
	}
}

func TestIsDuplicateUserEmailConflict_writeException(t *testing.T) {
	err := mongo.WriteException{
		WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "dup"}},
	}
	if !IsDuplicateUserEmailConflict(err) {
		t.Fatal("expected duplicate key from WriteException")
	}
}

func TestIsDuplicateUserEmailConflict_stringE11000(t *testing.T) {
	err := errors.New(`write exception: write errors: [E11000 duplicate key error collection: db.users index: email_1 dup key: { email: "a@b.com" }]`)
	if !ErrorDuplicatedData(err) {
		t.Fatal("expected duplicate key from error string")
	}
	if !IsDuplicateUserEmailConflict(err) {
		t.Fatal("expected duplicate user email conflict from error string")
	}
}
