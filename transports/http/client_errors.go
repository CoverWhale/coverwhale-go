package http

import (
	"fmt"
)

// AppError holds information about an application error
type ClientError struct {
	Status  int
	Details string
}

// ClientError is an interface to determine whether the error from the handler is a server or client error

// Error fulfills the error interface
func (c *ClientError) Error() string {
	return c.Details
}

// Body formats the application error for the caller
func (c *ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, c.Details))
}

func (c ClientError) As(target error) bool {
	_, ok := target.(*ClientError)
	return ok
}

// NewAppError creates a new application error
func NewClientError(err error, status int) error {
	return &ClientError{
		Status:  status,
		Details: err.Error(),
	}
}
