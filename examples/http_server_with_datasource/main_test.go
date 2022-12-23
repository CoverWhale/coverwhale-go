package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockDS struct {
	Data string
}

// GetData satisfies the datastore methods
func (m MockDS) GetData(id string) string {
	return m.Data
}

func TestExampleGet(t *testing.T) {
	// instantiate the test datastore
	ds := MockDS{
		Data: "testing",
	}
	// define the app and set the datastore
	a := App{
		DS: ds,
	}

	// set up a test table for our tests
	tt := []struct {
		name    string
		expect  string
		handler http.HandlerFunc
	}{
		{
			name:    "normal test",
			expect:  ds.Data,
			handler: a.exampleGet,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/testing", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			v.handler.ServeHTTP(rr, req)

			if status := rr.Code; status != 200 {
				t.Errorf("expected 200 but got %d", status)
			}

			data, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(data) != v.expect {
				t.Errorf("expected %s but got %s", v.expect, string(data))
			}
		})
	}

}
