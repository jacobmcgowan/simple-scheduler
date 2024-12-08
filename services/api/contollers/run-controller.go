package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
)

type RunController struct {
	runRepo repositories.RunRepository
}

func (cont RunController) Browse(ctx *gin.Context, filter dtos.RunFilter) {
	if runs, err := cont.runRepo.Browse(filter); err == nil {
		ctx.JSON(http.StatusOK, runs)
	} else {
		ctx.Error(err)
	}
}

func (cont RunController) Read(ctx *gin.Context, id string) {
	if run, err := cont.runRepo.Read(id); err == nil {
		ctx.JSON(http.StatusOK, run)
	} else {
		ctx.Error(err)
	}
}

func (cont RunController) Cancel(ctx *gin.Context, id string) {
	run, err := cont.runRepo.Read(id)
	if err != nil {
		ctx.Error(err)
		return
	}

	switch run.Status {
	case runStatuses.Cancelled:
	case runStatuses.Cancelling:
		ctx.Status(http.StatusNoContent)
	case runStatuses.Completed:
	case runStatuses.Failed:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "Run already finished",
		})
	case runStatuses.Pending:
	case runStatuses.Running:
		runUpdate := dtos.RunUpdate{
			Status: common.Undefinable[runStatuses.RunStatus]{
				Value:   runStatuses.Cancelling,
				Defined: true,
			},
		}

		if err := cont.runRepo.Edit(id, runUpdate); err == nil {
			ctx.Status(http.StatusNoContent)
		} else {
			ctx.Error(err)
		}
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "Invalid value",
		})
	}
}
