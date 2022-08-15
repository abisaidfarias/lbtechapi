package models

import "github.com/dgrijalva/jwt-go"

type AuthClaim struct {
	ID         string `json:"id"`
	CompanyID  string
	IsInternal string
	jwt.StandardClaims
}
