package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/validators"
)

func RegisterControllers(router *gin.Engine, jobRepo repositories.JobRepository, runRepo repositories.RunRepository) {
	api := router.Group("/api")

	status := api.Group("/status")
	status.GET("", func(ctx *gin.Context) {
		cont := StatusController{}
		cont.Get(ctx)
	})

	jobs := api.Group("/jobs")
	jobs.GET("", func(ctx *gin.Context) {
		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Browse(ctx)
	})
	jobs.GET("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobController{
			jobRepo: jobRepo,
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

		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Add(ctx, job)
	})
	jobs.PATCH("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")

		var jobUpdate dtos.JobUpdate
		if err := ctx.ShouldBindJSON(&jobUpdate); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Edit(ctx, name, jobUpdate)
	})
	jobs.DELETE("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Delete(ctx, name)
	})

	runs := api.Group("/runs")
	runs.GET("", func(ctx *gin.Context) {
		var filter dtos.RunFilter
		if err := ctx.ShouldBind(&filter); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if filter.Status != nil && !validators.ValidateRunStatus(string(*filter.Status), true) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "Invalid run status",
			})
		}

		cont := RunController{
			runRepo: runRepo,
		}
		cont.Browse(ctx, filter)
	})
	runs.GET("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		cont := RunController{
			runRepo: runRepo,
		}
		cont.Read(ctx, id)
	})
	runs.GET("/:id/cancel", func(ctx *gin.Context) {
		id := ctx.Param("id")
		cont := RunController{
			runRepo: runRepo,
		}
		cont.Cancel(ctx, id)
	})
}
