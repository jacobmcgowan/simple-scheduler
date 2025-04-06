package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/validators"
	ginoauth2 "github.com/zalando/gin-oauth2"
	"github.com/zalando/gin-oauth2/zalando"
)

func RegisterControllers(router *gin.Engine, jobRepo repositories.JobRepository, runRepo repositories.RunRepository) {
	api := router.Group("/api")

	status := api.Group("/status")
	status.GET("", func(ctx *gin.Context) {
		cont := StatusController{}
		cont.Get(ctx)
	})

	jobsReadScope := ginoauth2.Auth(zalando.ScopeAndCheck("jobs read", "jobs:read"), zalando.OAuth2Endpoint)
	jobsWriteScope := ginoauth2.Auth(zalando.ScopeAndCheck("jobs write", "jobs:write"), zalando.OAuth2Endpoint)
	jobs := api.Group("/jobs")
	jobs.GET("", func(ctx *gin.Context) {
		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Browse(ctx)
	}).Use(jobsReadScope)
	jobs.GET("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Read(ctx, name)
	}).Use(jobsReadScope)
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
	}).Use(jobsWriteScope)
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
	}).Use(jobsWriteScope)
	jobs.DELETE("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		cont := JobController{
			jobRepo: jobRepo,
		}
		cont.Delete(ctx, name)
	}).Use(jobsWriteScope)

	runsReadScope := ginoauth2.Auth(zalando.ScopeAndCheck("runs read", "runs:read"), zalando.OAuth2Endpoint)
	runsWriteScope := ginoauth2.Auth(zalando.ScopeAndCheck("runs write", "runs:write"), zalando.OAuth2Endpoint)
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
	}).Use(runsReadScope)
	runs.GET("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		cont := RunController{
			runRepo: runRepo,
		}
		cont.Read(ctx, id)
	}).Use(runsReadScope)
	runs.GET("/:id/cancel", func(ctx *gin.Context) {
		id := ctx.Param("id")
		cont := RunController{
			runRepo: runRepo,
		}
		cont.Cancel(ctx, id)
	}).Use(runsWriteScope)
}
