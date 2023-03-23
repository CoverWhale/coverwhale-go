package grist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GristTest struct {
	data     []byte
	response []byte
	document string
	key      string
}

func getHandler(gt GristTest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if gt.document == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if gt.key == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		w.Write(gt.data)
	}
}

func postHandler(gt GristTest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data json.RawMessage

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !bytes.Equal(gt.data, data) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		w.Write(gt.response)
	}
}

func TestGetDocument(t *testing.T) {
	tt := []struct {
		name     string
		key      string
		document string
		expected []byte
		err      error
	}{
		{name: "document request", key: "test", document: "document1", expected: []byte(`[{"id": 1, "name": "test}]`)},
		{name: "no api key", document: "document1", err: fmt.Errorf("%s", http.StatusText(http.StatusUnauthorized))},
		{name: "no document", document: "", err: fmt.Errorf("%s", http.StatusText(http.StatusNotFound))},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			gt := GristTest{
				data:     v.expected,
				document: v.document,
				key:      v.key,
			}

			s := httptest.NewServer(getHandler(gt))
			defer s.Close()

			c := NewClient(
				SetURL(s.URL),
				SetAPIKey(v.key),
			)

			res, err := c.GetDocument(v.document)
			if v.err != nil && v.err == nil {
				t.Errorf("expected no errors but got %v", err)
			}

			if !bytes.Equal(res, v.expected) {
				t.Errorf("expected %s but got %s", string(v.expected), string(res))
			}

		})
	}
}

func TestPostTable(t *testing.T) {
	tt := []struct {
		name     string
		data     []byte
		expected []byte
		err      error
	}{
		{name: "normal post", data: []byte(`{"records": [{"fields": {"test": 1}}]}`), expected: []byte(`{"records": [{"id": 1}]}`)},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			gt := GristTest{
				data:     v.data,
				response: v.expected,
			}

			s := httptest.NewServer(postHandler(gt))
			defer s.Close()

			c := NewClient(
				SetURL(s.URL),
			)

			res, err := c.CreateRecord("test", "test", bytes.NewReader(v.data))
			if err != nil && v.err == nil {
				t.Errorf("expected no errors but got %v", err)
			}

			if !bytes.Equal(res, v.expected) {
				t.Errorf("expected %s but got %s", string(v.expected), string(res))
			}

		})
	}
}
