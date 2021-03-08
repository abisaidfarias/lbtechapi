package util

import "errors"

// ErrorInvalidCredentials Invalid credentials"
var ErrorInvalidCredentials = errors.New("Invalid credentials")

// DATABASE ERROR

// ErrorDatabaseContection Unable to connect to database
var ErrorDatabaseContection = errors.New("Unable to connect to database")

// ErrorInQuery an error has occurred executing a query
var ErrorInQuery = errors.New("an error has occurred executing a query")
