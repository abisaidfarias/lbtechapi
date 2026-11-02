package utils

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrorInvalidCredentials Invalid credentials"
var ErrorInvalidCredentials = errors.New("invalid credentials")

// ErrorInvalidURLParams invalid url params
var ErrorInvalidURLParams = errors.New("invalid url params")

// DATABASE ERRORS

// ErrorResourceNotFound Unable to find the resource
var ErrorResourceNotFound = errors.New("resource not found")

// ErrorDatabaseContection Unable to connect to database
var ErrorDatabaseContection = errors.New("unable to connect to database")

// ErrorDuplicated duplicated data
var ErrorDuplicated = errors.New("duplicated data when creating a new document")

// ErrorDuplicateUserEmail is returned when the users collection unique email index rejects an insert.
var ErrorDuplicateUserEmail = errors.New("a user with this email address already exists")

// ErrorUserCreatedInviteFailed: user row was saved but the invitation email could not be sent.
var ErrorUserCreatedInviteFailed = errors.New("user created but invitation email could not be sent")

var ErrorUserNotPendingInvite = errors.New("invitation can only be resent for users who have not set their password yet")

var ErrorInviteResendRateLimited = errors.New("please wait before resending the invitation")

// ErrorInQuery an error has occurred executing a query
var ErrorInQuery = errors.New("an error has occurred executing a query")

var ErrorInvalidLineFormat = errors.New("invalid line format, invalid attributes amount or invalid category or invalid TestCode")

// ErrorUpgradeNotAllowed upgrade endpoint can only be used when no users exist
var ErrorUpgradeNotAllowed = errors.New("upgrade endpoint can only be used to create the first user. Users already exist in the system")

var ErrorPasswordNotSet = errors.New("password_not_set")
var ErrorInvalidPasswordToken = errors.New("invalid or expired password token")
var ErrorDeprecatedPasswordEndpoint = errors.New("deprecated endpoint; use forgot password or change your own password while signed in")

func writeErrorsHaveDuplicateCode(writeErrors mongo.WriteErrors) bool {
	for _, we := range writeErrors {
		if we.Code == 11000 {
			return true
		}
	}
	return false
}

// ErrorDuplicatedData reports MongoDB duplicate key errors (E11000), including wrapped driver types.
func ErrorDuplicatedData(err error) bool {
	if err == nil {
		return false
	}
	var merr mongo.WriteException
	if errors.As(err, &merr) && writeErrorsHaveDuplicateCode(merr.WriteErrors) {
		return true
	}
	var berr mongo.BulkWriteException
	if errors.As(err, &berr) {
		for _, we := range berr.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	// Some driver/server paths surface duplicate keys only in the error string.
	return strings.Contains(err.Error(), "E11000") || strings.Contains(err.Error(), "duplicate key")
}

// IsDuplicateUserEmailConflict is true when the error chain indicates a duplicate user email.
func IsDuplicateUserEmailConflict(err error) bool {
	if errors.Is(err, ErrorDuplicateUserEmail) {
		return true
	}
	return ErrorDuplicatedData(err)
}
