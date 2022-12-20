package http

import (
	"fmt"
	"runtime"
	"strings"
)

// AppError holds information about an application error
type AppError struct {
	status  int
	details string
}

// ClientError is an interface to determine whether the error from the handler is a server or client error
type ClientError interface {
	Error() string
	Body() []byte
	Status() int
}

// Error fulfills the error interface
func (a *AppError) Error() string {
	return a.details
}

// Body formats the application error for the caller
func (a *AppError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"`, a.details))
}

// Status is the HTTP status code of the error
func (a *AppError) Status() int {
	return a.status
}

// NewAppError creates a new application error
func NewAppError(err error, status int) error {
	return &AppError{
		status:  status,
		details: err.Error(),
	}
}

// GetCaller is a helper function to get the function name to provide context for an error
func GetCaller() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "function name unknown"
	}

	funcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	return funcName[len(funcName)-1]
}
