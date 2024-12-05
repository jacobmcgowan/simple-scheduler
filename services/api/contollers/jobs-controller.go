package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type JobsController struct {
	jobsRepo repositories.JobRepository
}

func (cont JobsController) Browse(ctx *gin.Context) {
	if jobs, err := cont.jobsRepo.Browse(); err == nil {
		ctx.JSON(http.StatusOK, jobs)
	} else {
		ctx.Error(err)
	}
}

func (cont JobsController) Read(ctx *gin.Context, name string) {
	if job, err := cont.jobsRepo.Read(name); err == nil {
		ctx.JSON(http.StatusOK, job)
	} else {
		ctx.Error(err)
	}
}

func (cont JobsController) Edit(ctx *gin.Context, name string, jobUpdate dtos.JobUpdate) {
	if err := cont.jobsRepo.Edit(name, jobUpdate); err == nil {
		ctx.Status(http.StatusNoContent)
	} else {
		ctx.Error(err)
	}
}

func (cont JobsController) Add(ctx *gin.Context, job dtos.Job) {
	if name, err := cont.jobsRepo.Add(job); err == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"name": name,
		})
	} else {
		ctx.Error(err)
	}
}

func (cont JobsController) Delete(ctx *gin.Context, name string) {
	if err := cont.jobsRepo.Delete(name); err == nil {
		ctx.Status(http.StatusNoContent)
	} else {
		ctx.Error(err)
	}
}
