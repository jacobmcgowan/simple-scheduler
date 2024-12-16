package responseHelpers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
)

func RespondWithError(ctx *gin.Context, err error) {
	var notFoundErr *repositoryErrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		ctx.Status(http.StatusNotFound)
	} else {
		ctx.Error(err)
	}
}
