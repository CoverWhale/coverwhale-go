package errors

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestClientErrorBody(t *testing.T) {
	tests := []struct {
		name         string
		clientError  ClientError
		expectedJSON string
	}{
		{
			name: "simple error with details only",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}]}`,
		},
		{
			name: "error with metadata",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
			).WithMetadataErrors(
				ErrorWithMetadata{
					Code:    "CWGEN1",
					Message: "Invalid input",
					Type:    "api",
					Level:   "warning",
				},
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}]}`,
		},
		{
			name: "error with params",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
				WithAdditionalParams(map[string]any{
					"field":  "email",
					"reason": "invalid format",
				}),
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}], "params": {"field": "email", "reason": "invalid format"}}`,
		},
		{
			name: "error with metadata and params",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
				WithAdditionalParams(map[string]any{
					"field":  "email",
					"reason": "invalid format",
				}),
			).WithMetadataErrors(
				ErrorWithMetadata{
					Code:    "CWGEN1",
					Message: "Invalid input",
					Type:    "api",
					Level:   "warning",
				},
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}], "params": {"field": "email", "reason": "invalid format"}}`,
		},
		{
			name: "error with submission id in params",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
				WithAdditionalParams(map[string]any{
					"submissionID": "2vjHqgXBWcptSeNzKn5CDeFp32L",
				}),
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}], "params": {"submissionID": "2vjHqgXBWcptSeNzKn5CDeFp32L"}}`,
		},
		{
			name: "error with nested params",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
				WithAdditionalParams(map[string]any{
					"validation": map[string]any{
						"field":  "email",
						"reason": "invalid format",
					},
				}),
			),
			expectedJSON: `{"errors": [{"code": "CWGEN1", "message": "Invalid input", "type": "api", "level": "warning"}], "params": {"validation": {"field": "email", "reason": "invalid format"}}}`,
		},
		{
			name: "error with custom metadata",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
			).WithMetadataErrors(
				ErrorWithMetadata{
					Code:    "CUSTOM1",
					Message: "Custom error message",
					Type:    "validation",
					Level:   "error",
				},
			),
			expectedJSON: `{"errors": [{"code": "CUSTOM1", "message": "Custom error message", "type": "validation", "level": "error"}]}`,
		},
		{
			name: "error with custom metadata and params",
			clientError: NewClientError(
				fmt.Errorf("Invalid input"),
				400,
				WithAdditionalParams(map[string]any{
					"field":  "email",
					"reason": "invalid format",
				}),
			).WithMetadataErrors(
				ErrorWithMetadata{
					Code:    "CUSTOM1",
					Message: "Custom error message",
					Type:    "validation",
					Level:   "error",
				},
			),
			expectedJSON: `{"errors": [{"code": "CUSTOM1", "message": "Custom error message", "type": "validation", "level": "error"}], "params": {"field": "email", "reason": "invalid format"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.clientError.Body()

			var expected, actual interface{}
			if err := json.Unmarshal([]byte(tt.expectedJSON), &expected); err != nil {
				t.Fatalf("failed to unmarshal expected JSON: %v", err)
			}
			if err := json.Unmarshal(got, &actual); err != nil {
				t.Fatalf("failed to unmarshal actual JSON: %v", err)
			}

			expectedJSON, err := json.Marshal(expected)
			if err != nil {
				t.Fatalf("failed to marshal expected JSON: %v", err)
			}
			actualJSON, err := json.Marshal(actual)
			if err != nil {
				t.Fatalf("failed to marshal actual JSON: %v", err)
			}
			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("Body() = %v, want %v", string(actualJSON), string(expectedJSON))
			}
		})
	}
}
