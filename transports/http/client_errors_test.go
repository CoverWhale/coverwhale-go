package http

import (
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	tt := []struct {
		name string
		err  ClientError
		want string
	}{
		{name: "simple", err: ClientError{Details: "test", Status: 400}, want: "test"},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			if v.err.Error() != v.want {
				t.Errorf("expected %s but got %s", v.want, v.err.Error())
			}

			wantBody := fmt.Sprintf(`{"error": "%s"}`, v.want)
			if string(v.err.Body()) != wantBody {
				t.Errorf("expected %s, but got %s", wantBody, string(v.err.Body()))
			}
		})
	}
}
