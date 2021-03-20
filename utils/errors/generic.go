package utils

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrorInvalidCredentials Invalid credentials"
var ErrorInvalidCredentials = errors.New("Invalid credentials")

// ErrorInvalidURLParams invalid url params
var ErrorInvalidURLParams = errors.New("Invalid url params")

// DATABASE ERRORS

// ErrorResourceNotFound Unable to find the resource
var ErrorResourceNotFound = errors.New("Resource not found")

// ErrorDatabaseContection Unable to connect to database
var ErrorDatabaseContection = errors.New("Unable to connect to database")

// ErrorDuplicated duplicated data
var ErrorDuplicated = errors.New("Duplicated data when creating a new document")
var ErrorDuplicatedUser = errors.New("this user is duplicated, please verify the data provided ")
var ErrorDuplicatedTestCode = errors.New("The test code is duplicated, please verify the data provided")
var ErrorDuplicatedTestPlan = errors.New("The test plan is duplicated, please verify the data provided")

// ErrorInQuery an error has occurred executing a query
var ErrorInQuery = errors.New("An error has occurred executing a query")

func ErrorDuplicatedData(err error) bool {
	var merr mongo.WriteException
	merr = err.(mongo.WriteException)
	errCode := merr.WriteErrors[0].Code
	if errCode == 11000 {
		return true
	}
	return false
}
