package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type authHeader struct {
	Token string `header:"Authorization" binding:"required"`
}

var (
	userRepository repositories.IUserRepository = repositories.NewUserRepository()
	userService    services.IUserService        = services.NewUserService(userRepository)
)

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

		// TODO remove database query to get user
		user, err := userService.GetByID(claims.ID)
		// TODO  get payload from token to gin ctx

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		ctx.Set("user", user)

		ctx.Next()
	}
}

func validateToken(tokenString string) (*services.AuthClaims, error) {

	claims := &services.AuthClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return services.JWTKey, nil
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
