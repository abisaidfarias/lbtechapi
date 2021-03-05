package models

import "errors"

// ErrorInvalidCredentials error
var ErrorInvalidCredentials = errors.New("Invalid credentials")

// ErrorDatabaseContection error
var ErrorDatabaseContection = errors.New("Unable to connect to database")
