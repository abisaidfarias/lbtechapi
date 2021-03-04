package models

import "errors"

// ErrorInvalidCredentials error
var ErrorInvalidCredentials = errors.New("invalid credentials")

// ErrorDatabaseContection error
var ErrorDatabaseContection = errors.New("unable to connect to database")
