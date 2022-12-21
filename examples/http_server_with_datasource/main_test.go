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

func (m MockDS) GetData(id string) string {
	return m.Data
}

func TestExampleGet(t *testing.T) {
	ds := MockDS{
		Data: "testing",
	}
	a := App{
		DS: ds,
	}

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
