package middlewares

import (
	"net/http"
	"strings"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var (
	authService services.AuthService
)

type authHeader struct {
	jwt string `header:"Authorization"`
}

// AuthMiddleware is the jwt middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := authHeader{}

		// getting header to stract the jwt
		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}

		token := h.jwt

		idTokenHeader := strings.Split(h.jwt, "Bearer ")

		if len(idTokenHeader) < 2 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer is required"})
			ctx.Abort()
			return
		}

		claims, err := validateToken(token)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		id := claims.ID
		user, err := authService.GetUserByID(id)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}

		ctx.Set("user", *user)

		ctx.Next()
	}
}

func validateToken(tokenString string) (*services.AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return services.JWTKey, nil
	})

	if err == nil && token.Valid {
		claims, _ := token.Claims.(services.AuthClaims)
		return &claims, nil
	}

	return nil, err
}
