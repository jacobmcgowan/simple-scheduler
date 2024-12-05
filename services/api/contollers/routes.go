package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	jsonPatchMerge "github.com/jacobmcgowan/simple-scheduler/services/api/json-patch-merge"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

func RegisterControllers(router *gin.Engine, jobsRepo repositories.JobRepository) {
	api := router.Group("/api")

	status := api.Group("/status")
	status.GET("", func(ctx *gin.Context) {
		cont := StatusController{}
		cont.Get(ctx)
	})

	jobs := api.Group("/jobs")
	jobs.GET("", func(ctx *gin.Context) {
		cont := JobsController{
			jobsRepo: jobsRepo,
		}
		cont.Browse(ctx)
	})
	jobs.GET("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobsController{
			jobsRepo: jobsRepo,
		}
		cont.Read(ctx, name)
	})
	jobs.POST("", func(ctx *gin.Context) {
		var job dtos.Job
		if err := ctx.ShouldBindJSON(&job); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		cont := JobsController{
			jobsRepo: jobsRepo,
		}
		cont.Add(ctx, job)
	})
	jobs.PATCH("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		jsonData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		jobUpdate, err := jsonPatchMerge.PatchJobUpdate(jsonData)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		cont := JobsController{
			jobsRepo: jobsRepo,
		}
		cont.Edit(ctx, name, jobUpdate)
	})
	jobs.DELETE("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobsController{
			jobsRepo: jobsRepo,
		}
		cont.Delete(ctx, name)
	})
}
