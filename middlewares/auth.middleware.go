package middlewares

import (
	"fmt"
	// "net/http"
	// "strings"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/dgrijalva/jwt-go"
	// "github.com/gin-gonic/gin"
)

var (
	authService services.IAuthService
)

type authHeader struct {
	IDToken string `header:"Authorization"`
}

// idTokenCustomClaims holds structure of jwt claims of idToken
type idTokenCustomClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// AuthMiddleware is the jwt middleware
// func AuthMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		h := authHeader{}

// 		// getting header to stract the jwt
// 		if err := ctx.ShouldBindHeader(&h); err != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		}

// 		idTokenHeader := strings.Split(h.IDToken, "Bearer ")

// 		if len(idTokenHeader) < 2 {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer is required"})
// 			ctx.Abort()
// 			return
// 		}

// 		claims, err := validateToken(idTokenHeader[1])

// 		if err != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 			ctx.Abort()
// 			return
// 		}

// 		user, err := authService.GetUserByID(claims.ID)

// 		if err != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 			ctx.Abort()
// 			return
// 		}

// 		ctx.Set("user", user)

// 		ctx.Next()
// 	}
// }

func validateToken(tokenString string) (*idTokenCustomClaims, error) {

	claims := &idTokenCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return services.JWTKey, nil
	})

	// For now we'll just return the error and handle logging in service level
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("ID token is invalid")
	}

	claims, ok := token.Claims.(*idTokenCustomClaims)

	if !ok {
		return nil, fmt.Errorf("ID token valid but couldn't parse claims")
	}

	return claims, nil

}
