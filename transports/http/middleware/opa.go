package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/opa"
	"github.com/CoverWhale/logr"
)

type Query struct {
	OperationName string    `json:"operationName,omitempty"`
	Variables     Variables `json:"variables"`
}

type Variables struct {
	Data any `json:"data"`
}

// ValidationFunc is used to map the data in the incoming request to an OPA request.
// It's on the caller of the service to define the data in the OPA request and
// what package is called in OPA.
type ValidationFunc func([]byte) (opa.OPARequest, error)

// StandardValidator calls the central OPA server with the full cw validation pacakge
func StandardValidator(h http.Handler) http.Handler {
	url := fmt.Sprintf("%s/v1/data/cw", opa.CentralOPA)

	fn := func(w http.ResponseWriter, r *http.Request) {

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logr.Errorf("error reading request body: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// create two copies of request body since reading the body drains the reader
		opaBody := io.NopCloser(bytes.NewBuffer(buf))
		dataBody := io.NopCloser(bytes.NewBuffer(buf))

		if !validate(w, opaBody, url) {
			return
		}

		// reset the request body to the copied reader
		r.Body = dataBody

		h.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)
}

// CustomValidtor allows for calling a specific OPA instance and defining the package name
// For example, to call a sidecar instead of the central OPA server and to only call
// the cw/underwriting package instead of the full cw package.
func CustomValidator(h http.Handler, customValidator ValidationFunc, url opa.OPAURL, pkg string) http.Handler {
	endpoint := fmt.Sprintf("%s/v1/data/%s", url, pkg)

	fn := func(w http.ResponseWriter, r *http.Request) {

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logr.Errorf("error reading request body: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// create two copies of request body since reading the body drains the reader
		dataBody := io.NopCloser(bytes.NewBuffer(buf))

		opaRequest, err := customValidator(buf)
		if err != nil {
			logr.Errorf("error from custom validation func: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		requestData, err := json.Marshal(opaRequest)
		if err != nil {
			logr.Errorf("error marshaling input data: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !validate(w, bytes.NewReader(requestData), endpoint) {
			return
		}

		// reset the request body to the copied reader
		r.Body = dataBody

		h.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)
}

// GraphQLCustomValidator is like the other CustomValidator function but for GraphQL requests. It passes through data if the
// query type is an introspection so that schema requests aren't validated by OPA
func GraphQLCustomValidator(handler http.Handler, customValidator ValidationFunc, url opa.OPAURL, pkg string) http.Handler {
	endpoint := fmt.Sprintf("%s/v1/data/%s", url, pkg)

	fn := func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logr.Errorf("error reading request body: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// create two copies of request body since reading the body drains the reader
		queryBody := io.NopCloser(bytes.NewBuffer(buf))
		dataBody := io.NopCloser(bytes.NewBuffer(buf))

		data, err := io.ReadAll(queryBody)
		if err != nil {
			return
		}
		defer queryBody.Close()

		var q Query

		if err := json.Unmarshal(data, &q); err != nil {
			logr.Errorf("error unmarshaling query data: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// return introspection queries directly for schema lookups.
		if q.OperationName == "IntrospectionQuery" {
			r.Body = dataBody
			handler.ServeHTTP(w, r)
			return
		}

		opaData, err := json.Marshal(q.Variables.Data)
		if err != nil {
			return
		}

		opaRequest, err := customValidator(opaData)
		if err != nil {
			logr.Errorf("error from custom validation func: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		requestData, err := json.Marshal(opaRequest)
		if err != nil {
			logr.Errorf("error marshaling input data: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !validate(w, bytes.NewReader(requestData), endpoint) {
			return
		}

		// reset the request body to the copied reader
		r.Body = dataBody

		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func validate(w http.ResponseWriter, r io.Reader, url string) bool {
	var or opa.OPAResponse

	data, err := io.ReadAll(r)
	if err != nil {
		logr.Errorf("error decoding OPA request data: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		logr.Errorf("error building OPA request: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}

	if resp.StatusCode != 200 {
		http.Error(w, http.StatusText(resp.StatusCode), resp.StatusCode)
		return false
	}

	if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
		logr.Errorf("error decoding response from OPA: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}

	if !or.Result.Allow {
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(or.Result); err != nil {
			logr.Errorf("error encoding OPA unauthorized response: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return false
		}
		return false
	}

	return true
}
