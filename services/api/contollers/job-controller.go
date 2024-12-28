package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	responseHelpers "github.com/jacobmcgowan/simple-scheduler/services/api/response-helpers"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type JobController struct {
	jobRepo repositories.JobRepository
}

func (cont JobController) Browse(ctx *gin.Context) {
	if jobs, err := cont.jobRepo.Browse(); err == nil {
		ctx.JSON(http.StatusOK, jobs)
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont JobController) Read(ctx *gin.Context, name string) {
	if job, err := cont.jobRepo.Read(name); err == nil {
		ctx.JSON(http.StatusOK, job)
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont JobController) Edit(ctx *gin.Context, name string, jobUpdate dtos.JobUpdate) {
	if err := cont.jobRepo.Edit(name, jobUpdate); err == nil {
		ctx.Status(http.StatusNoContent)
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont JobController) Add(ctx *gin.Context, job dtos.Job) {
	if name, err := cont.jobRepo.Add(job); err == nil {
		ctx.JSON(http.StatusCreated, gin.H{
			"name": name,
		})
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont JobController) Delete(ctx *gin.Context, name string) {
	if err := cont.jobRepo.Delete(name); err == nil {
		ctx.Status(http.StatusNoContent)
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}
