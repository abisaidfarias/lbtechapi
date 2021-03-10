package middlewares

import (
	"os"
	"fmt"
	"net/http"
	"strings"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type authHeader struct {
	Token string `header:"Authorization" binding:"required"`
}

// AuthMiddleware is the jwt middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := authHeader{}

		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		idTokenHeader := strings.Split(h.Token, "Bearer ")

		if len(idTokenHeader) < 2 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer is required in authorization token"})
			ctx.Abort()
			return
		}

		claims, err := validateToken(idTokenHeader[1])

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.ID)
		// TODO add this after company relation
		// ctx.Set("companyID", claims.companyID)

		ctx.Next()
	}
}

func validateToken(tokenString string) (*services.AuthClaims, error) {

	var JWTKey = []byte(os.Getenv("SECRET_KEY"))
	claims := &services.AuthClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("ID token is invalid")
	}

	claims, ok := token.Claims.(*services.AuthClaims)

	if !ok {
		return nil, fmt.Errorf("ID token valid but couldn't parse claims")
	}

	return claims, nil
}
