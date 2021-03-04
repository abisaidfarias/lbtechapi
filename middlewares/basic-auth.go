package middlewares

import "github.com/gin-gonic/gin"

//BasicAuth function
func BasicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		"abisaid": "@Mipassword123",
	})
}
