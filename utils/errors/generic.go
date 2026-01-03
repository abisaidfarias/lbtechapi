package utils

import (
	"errors"

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

// ErrorInQuery an error has occurred executing a query
var ErrorInQuery = errors.New("an error has occurred executing a query")

var ErrorInvalidLineFormat = errors.New("invalid line format, invalid attributes amount or invalid category or invalid TestCode")

// ErrorUpgradeNotAllowed upgrade endpoint can only be used when no users exist
var ErrorUpgradeNotAllowed = errors.New("upgrade endpoint can only be used to create the first user. Users already exist in the system")

func ErrorDuplicatedData(err error) bool {
	var merr mongo.WriteException
	// Usar errors.As para manejar errores envueltos con fmt.Errorf
	if errors.As(err, &merr) {
		if len(merr.WriteErrors) > 0 {
			errCode := merr.WriteErrors[0].Code
			if errCode == 11000 {
				return true
			}
		}
	}
	return false
}
