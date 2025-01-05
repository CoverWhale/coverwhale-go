// Copyright 2025 Sencillo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"
)

// ClientError represents a non-server error
type ClientError struct {
	// Status is the status code to be returned
	Status int

	// Details are a nicely formatted client error
	Details string

	//DetailedError is the actual error to be logged
	DetailedError error
}

type ClientErrorOpt func(*ClientError)

func (c ClientError) Error() string {
	return c.Details
}

func (c ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": %q}`, c.Details))
}

func (c ClientError) Code() int {
	return c.Status
}

func (c ClientError) LoggedError() string {
	return c.DetailedError.Error()
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

func WithDetailedError(err error) ClientErrorOpt {
	return func(c *ClientError) {
		c.DetailedError = err
	}
}

func NewClientError(err error, code int, opts ...ClientErrorOpt) ClientError {
	ce := ClientError{
		Status:        code,
		Details:       err.Error(),
		DetailedError: err,
	}

	for _, v := range opts {
		v(&ce)
	}

	return ce
}
