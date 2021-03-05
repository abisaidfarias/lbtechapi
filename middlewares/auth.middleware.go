package middlewares

import (
	"net/http"

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

		claims, err := extractClaims(token)

		user, err := authService.GetUserByID(claims.ID)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}

		ctx.Set("user", *user)

		ctx.Next()
	}
}

func extractClaims(token string) (*services.AuthClaims, error) {
	claims := services.AuthClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return services.JwtKey, nil
	})

	if err != nil {
		return nil, err
	}
	return &claims, nil
}
