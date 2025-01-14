package httpHelpers

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	params map[string]string
}

func NewQueryBuilder() QueryBuilder {
	return QueryBuilder{
		params: make(map[string]string),
	}
}

func (qb *QueryBuilder) Add(param string, value *string) {
	if value != nil {
		qb.params[param] = *value
	} else {
		delete(qb.params, param)
	}
}

func (qb QueryBuilder) String() string {
	var sb strings.Builder
	for param, val := range qb.params {
		if sb.Len() == 0 {
			sb.WriteString(fmt.Sprintf("?%s=%s", param, val))
		} else {
			sb.WriteString(fmt.Sprintf("&%s=%s", param, val))
		}
	}

	return sb.String()
}
