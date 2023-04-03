package grist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type GristTest struct {
	expected Data
	data     Data
	document string
	key      string
	filter   map[string]json.RawMessage
}

type Data struct {
	Records []Record `json:"records"`
}

type Record struct {
	ID     int     `json:"id"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Foo string `json:"foo"`
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

		if gt.filter != nil {
			gt.data.Records = gt.data.Records[:1]
		}

		if err := json.NewEncoder(w).Encode(gt.data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func postHandler(gt GristTest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data Data

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !reflect.DeepEqual(gt.data, data) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(gt.expected); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func TestGetDocument(t *testing.T) {
	tt := []struct {
		name     string
		key      string
		document string
		expected Data
		data     Data
		err      error
		filter   map[string]json.RawMessage
	}{
		{
			name: "document request",
			key:  "test", document: "document1",
			data: Data{
				Records: []Record{
					{
						ID: 1,
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
				},
			},
			expected: Data{
				Records: []Record{
					{
						ID: 1,
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
				},
			},
		},
		{
			name: "filtered document request",
			key:  "test", document: "document1",
			data: Data{
				Records: []Record{
					{
						ID: 1,
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
					{
						ID: 1,
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
				},
			},
			expected: Data{
				Records: []Record{
					{
						ID: 1,
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
				},
			},
			filter: map[string]json.RawMessage{
				"filter": json.RawMessage(`{"id": [1]}`),
			},
		},
		{name: "no api key", document: "document1", err: fmt.Errorf("%s", http.StatusText(http.StatusUnauthorized))},
		{name: "no document", document: "", err: fmt.Errorf("%s", http.StatusText(http.StatusNotFound))},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			gt := GristTest{
				expected: v.expected,
				document: v.document,
				key:      v.key,
				data:     v.data,
				filter:   v.filter,
			}

			s := httptest.NewServer(getHandler(gt))
			defer s.Close()

			c := NewClient(
				SetURL(s.URL),
				SetAPIKey(v.key),
			)

			res, err := c.GetFilteredRecords(v.document, "", v.filter)
			if v.err != nil && v.err == nil {
				t.Errorf("expected no errors but got %v", err)
			}

			// expected error here so we are good
			if v.err != nil && err != nil {
				return
			}

			var d Data
			if err := json.Unmarshal(res, &d); err != nil {
				t.Errorf("error unmarshaling: %v", err)
			}

			if !reflect.DeepEqual(d, v.expected) {
				t.Errorf("expected \n%#v\nbut got \n%#v", v.expected, d)
			}

		})
	}
}

func TestCreateRecord(t *testing.T) {
	tt := []struct {
		name     string
		data     Data
		expected Data
		err      error
	}{
		{
			name: "normal post",
			data: Data{
				Records: []Record{
					{
						Fields: []Field{
							{
								Foo: "test",
							},
						},
					},
				},
			},
			expected: Data{
				Records: []Record{
					{
						ID: 1,
					},
				},
			},
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			gt := GristTest{
				data:     v.data,
				expected: v.expected,
			}

			s := httptest.NewServer(postHandler(gt))
			defer s.Close()

			c := NewClient(
				SetURL(s.URL),
			)

			d, err := json.Marshal(v.data)
			if err != nil {
				t.Errorf("error marshaling: %v", err)
			}

			res, err := c.CreateRecord("test", "test", bytes.NewReader(d))
			if err != nil && v.err == nil {
				t.Errorf("expected no errors but got %v", err)
			}

			// expected error here so we are good
			if v.err != nil && err != nil {
				return
			}

			var respData Data
			if err := json.Unmarshal(res, &respData); err != nil {
				t.Errorf("error unmarshaling: %v", err)
			}

			if !reflect.DeepEqual(respData, v.expected) {
				t.Errorf("expected \n%#v\nbut got \n%#v", v.expected, respData)
			}

		})
	}
}
