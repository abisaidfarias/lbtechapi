package utils

import "errors"

// ErrorInvalidCredentials Invalid credentials"
var ErrorInvalidCredentials = errors.New("Invalid credentials")

// DATABASE ERRORS

// ErrorDatabaseContection Unable to connect to database
var ErrorDatabaseContection = errors.New("Unable to connect to database")

// TODO wrap error 11000

// ErrorDuplicated duplicated data
var ErrorDuplicated = errors.New("Duplicated data when creating a new document")

// ErrorInQuery an error has occurred executing a query
var ErrorInQuery = errors.New("An error has occurred executing a query")
