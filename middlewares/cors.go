package middlewares

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		environment := os.Getenv("ENVIRONMENT")

		// En producción, solo permitir orígenes específicos
		if environment == "prod" {
			allowedOrigins := []string{
				"https://lbonetrack.com",
				"https://www.lbonetrack.com",
			}

			// Verificar si el origen está en la lista permitida
			if isOriginAllowed(origin, allowedOrigins) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Si no está permitido, no se establece el header (el navegador bloqueará la petición)
		} else {
			// En desarrollo, permitir todos los orígenes para facilitar testing
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH,OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed verifica si el origen está en la lista de orígenes permitidos
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	// Normalizar el origen (remover trailing slash)
	origin = strings.TrimSuffix(origin, "/")

	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSuffix(allowed, "/")
		if origin == allowed {
			return true
		}
	}

	return false
}
