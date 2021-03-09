package controllers

import (
	"errors"
	"net/http"

	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/gin-gonic/gin"
)

func handleErrorResponse(ctx *gin.Context, err error) {
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrorInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		case errors.Is(err, utils.ErrorResourceNotFound), errors.Is(err, utils.ErrorInQuery), errors.Is(err, utils.ErrorDuplicated):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	return
}
