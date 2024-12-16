package controllers

import (
	"github.com/gin-gonic/gin"
)

type StatusController struct {
}

func (cont StatusController) Get(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "OK",
	})
}
