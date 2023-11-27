package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CoverWhale/logr"
)

const (
	SideCarOPA OPAURL = "http://localhost:8181"
	CentralOPA OPAURL = "http://opa.svc.cluster.local:8181"
)

type OPAURL string

type OPAResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Allow bool     `json:"allow"`
	Deny  []string `json:"deny,omitempty"`
}

type OPARequest struct {
	Input Input `json:"input"`
}

type Input struct {
	State        string    `json:"state"`
	BusinessType string    `json:"business_type"`
	Commodities  []string  `json:"commodities"`
	Drivers      []Driver  `json:"drivers"`
	Vehicles     []Vehicle `json:"vehicles"`
	Trailers     []Trailer `json:"trailers"`
}

type Driver struct {
	Name       string   `json:"name"`
	Experience int      `json:"experience"`
	Age        int      `json:"age"`
	AVDs       []string `json:"avds"`
}

type Vehicle struct {
	ID        string `json:"id"`
	BodyType  string `json:"body_type"`
	Class     int    `json:"class"`
	ModelYear int    `json:"model_year"`
	Amount    int    `json:"amount"`
}

type Trailer struct {
	ID          string `json:"id"`
	TrailerType string `json:"trailer_type"`
	ModelYear   int    `json:"model_year"`
	Amount      int    `json:"amount"`
}

func StandardValidator(h http.Handler) http.Handler {
	url := fmt.Sprintf("%s/v1/data/cw", SideCarOPA)

	fn := func(w http.ResponseWriter, r *http.Request) {

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logr.Errorf("error reading request body: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

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

func CustomValidator(h http.Handler, url OPAURL, pkg string) http.Handler {
	endpoint := fmt.Sprintf("%s/v1/data/%s", url, pkg)

	fn := func(w http.ResponseWriter, r *http.Request) {

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logr.Errorf("error reading request body: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create two copies of request body since reading the body drains the reader
		opaBody := io.NopCloser(bytes.NewBuffer(buf))
		dataBody := io.NopCloser(bytes.NewBuffer(buf))

		if !validate(w, opaBody, endpoint) {
			return
		}

		// reset the request body to the copied reader
		r.Body = dataBody

		h.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)
}

func validate(w http.ResponseWriter, r io.Reader, url string) bool {
	var or OPAResponse

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
