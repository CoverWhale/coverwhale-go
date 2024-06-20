package errors

import (
	"fmt"
)

// ClientError represents a non-server error
type ClientError struct {
	Status  int
	Details string
}

func (c ClientError) Error() string {
	return c.Details
}

func (c ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, c.Details))
}

func (c ClientError) Code() int {
	return c.Status
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

func NewClientError(err error, code int) ClientError {
	return ClientError{
		Status:  code,
		Details: err.Error(),
	}
}
