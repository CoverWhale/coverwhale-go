package errors

import (
	"fmt"
	"strings"
)

// ClientError represents a non-server error
type ClientError struct {
	// Status is the status code to be returned
	Status int

	// Details are a nicely formatted client error
	Details string

	//Objects is a slice of error objects
	ErrorsWithMetadata []ErrorWithMetadata

	//DetailedError is the actual error to be logged
	DetailedError error
}

type ErrorWithMetadata struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Level   string `json:"level"`
}

type ClientErrorOpt func(*ClientError)

func (c ClientError) Error() string {
	return c.Details
}

func (c ClientError) Body() []byte {
	format := `{"code": %q, "message": %q, "type": %q, "level": %q}`
	if c.ErrorsWithMetadata != nil {
		return []byte(fmt.Sprintf(`{"errors": [%s]}`, errorMetadataToFormattedString(format, c.ErrorsWithMetadata...)))
	}

	return []byte(fmt.Sprintf(`{"errors": [%q]}`, c.Details))
}

func (c ClientError) Code() int {
	return c.Status
}

func (c ClientError) LoggedError() string {
	format := `{code: %s, message: %s, type: %s, level: %s}`
	if c.ErrorsWithMetadata != nil {
		return errorMetadataToFormattedString(format, c.ErrorsWithMetadata...)
	}

	return c.DetailedError.Error()
}

func errorMetadataToFormattedString(format string, objects ...ErrorWithMetadata) string {
	var errors []string
	for _, v := range objects {
		errors = append(errors, fmt.Sprintf(format, v.Code, v.Message, v.Type, v.Level))
	}

	return strings.Join(errors, ", ")
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

func (c ClientError) WithMetadataErrors(objects ...ErrorWithMetadata) ClientError {
	c.ErrorsWithMetadata = objects

	return c
}

func WithDetailedError(err error) ClientErrorOpt {
	return func(c *ClientError) {
		c.DetailedError = err
	}
}

func NewClientError(err error, code int, opts ...ClientErrorOpt) ClientError {
	metadata := ErrorWithMetadata{
		Code:    "CWGEN100",
		Message: err.Error(),
		Type:    "api",
		Level:   "warning",
	}
	ce := ClientError{
		Status:             code,
		Details:            err.Error(),
		ErrorsWithMetadata: []ErrorWithMetadata{metadata},
		DetailedError:      err,
	}

	for _, v := range opts {
		v(&ce)
	}

	return ce
}
