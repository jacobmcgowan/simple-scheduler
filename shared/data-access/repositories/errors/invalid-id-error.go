package repositoryErrors

import "fmt"

type InvalidIdError struct {
	Value string
}

func (err *InvalidIdError) Error() string {
	return fmt.Sprintf("Invalid value for ID: %s", err.Value)
}
