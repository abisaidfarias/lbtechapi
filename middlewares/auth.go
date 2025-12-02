package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/models"
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
		ctx.Set("companyID", claims.CompanyID)

		ctx.Next()
	}
}

func validateToken(tokenString string) (*models.AuthClaim, error) {

	var JWTKey = []byte(config.GetValue("SECRET_KEY"))
	claims := &models.AuthClaim{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("ID token is invalid")
	}

	claims, ok := token.Claims.(*models.AuthClaim)

	if !ok {
		return nil, fmt.Errorf("ID token valid but couldn't parse claims")
	}

	return claims, nil
}
