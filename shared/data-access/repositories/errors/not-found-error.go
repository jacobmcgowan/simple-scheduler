package repositoryErrors

type NotFoundError struct {
	Message string
}

func (err *NotFoundError) Error() string {
	return err.Message
}
