package converters

import (
	"github.com/gin-gonic/gin"
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
)

func ParamToUndefinableString(params gin.Params, name string) common.Undefinable[string] {
	if param, hasParam := params.Get(name); hasParam {
		return common.Undefinable[string]{
			Value:   param,
			Defined: hasParam,
		}
	}

	return common.Undefinable[string]{
		Defined: false,
	}
}
