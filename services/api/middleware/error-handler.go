package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 && ctx.Writer.Status() == http.StatusOK {
			log.Printf("Error for request %s %s: %s", ctx.Request.Method, ctx.Request.RequestURI, ctx.Errors.String())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unexpected error occurred",
			})
		}
	}
}
