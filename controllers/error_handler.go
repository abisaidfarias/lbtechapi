package controllers

import (
	"errors"
	"net/http"

	apputils "github.com/abisaidfarias/lbtechapi/utils"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/gin-gonic/gin"
)

func respondDuplicateUserEmail(ctx *gin.Context) {
	ctx.JSON(http.StatusConflict, gin.H{
		"error": utils.ErrorDuplicateUserEmail.Error(),
		"code":  "duplicate_email",
	})
}

func handleErrorResponse(ctx *gin.Context, err error) {
	if err != nil {
		switch {
		// Business-rule errors (utils.NewValidationError) are client errors,
		// whichever handler forwards them here.
		case apputils.IsValidationError(err):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		case errors.Is(err, utils.ErrorInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		case errors.Is(err, utils.ErrorResourceNotFound), errors.Is(err, utils.ErrorInQuery), errors.Is(err, utils.ErrorDuplicated):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}
	}
}
