package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	responseHelpers "github.com/jacobmcgowan/simple-scheduler/services/api/response-helpers"
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
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont RunController) Read(ctx *gin.Context, id string) {
	if run, err := cont.runRepo.Read(id); err == nil {
		ctx.JSON(http.StatusOK, run)
	} else {
		responseHelpers.RespondWithError(ctx, err)
	}
}

func (cont RunController) Cancel(ctx *gin.Context, id string) {
	run, err := cont.runRepo.Read(id)
	if err != nil {
		responseHelpers.RespondWithError(ctx, err)
		return
	}

	switch run.Status {
	case runStatuses.Cancelled:
	case runStatuses.Cancelling:
		ctx.Status(http.StatusNoContent)
	case runStatuses.Completed:
	case runStatuses.Failed:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Run already finished",
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
			responseHelpers.RespondWithError(ctx, err)
		}
	default:
		ctx.Error(fmt.Errorf("run in unexpected status %s", run.Status))
	}
}
