// Copyright 2023 Cover Whale Insurance Solutions Inc.
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/opa"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/CoverWhale/coverwhale-go/transports/http/middleware"
	"github.com/CoverWhale/logr"
)

type Request struct {
	Operation   string    `json:"operation"`
	Commodities []string  `json:"commodities"`
	Vehicles    []Vehicle `json:"vehicles"`
}

type Vehicle struct {
	VIN      string `json:"vin"`
	BodyType string `json:"body_type"`
	Class    int    `json:"class"`
	Amount   int    `json:"amount"`
}

func getRoutes(l *logr.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodPost,
			Path:    "/test",
			Handler: http.HandlerFunc(test),
		},
		{
			Method:  http.MethodPost,
			Path:    "/test-custom",
			Handler: middleware.CustomValidator(http.HandlerFunc(test), opaValidate, opa.SideCarOPA, "cw/underwriting"),
		},
		{
			Method:  http.MethodPost,
			Path:    "/test-custom-decision",
			Handler: middleware.CustomValidator(http.HandlerFunc(test), opaValidateDecision, opa.SideCarOPA, "cw/underwriting/decision"),
		},
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logr.Errorf("error decoding app data: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var total int

	for _, v := range req.Vehicles {
		total += v.Amount
	}

	resp := fmt.Sprintf("they're worth $%d", total)
	w.Write([]byte(resp))
}

// custom validation function to send to OPA.
func opaValidate(data []byte) (opa.OPARequest, error) {
	var req Request

	if err := json.Unmarshal(data, &req); err != nil {
		return opa.OPARequest{}, err
	}

	var vehicles []opa.Vehicle

	for _, v := range req.Vehicles {
		vehicles = append(vehicles, opa.Vehicle{
			ID:       v.VIN,
			BodyType: v.BodyType,
			Class:    v.Class,
			Amount:   v.Amount,
		})
	}

	return opa.OPARequest{
		Input: opa.Input{
			Operation:   req.Operation,
			Commodities: req.Commodities,
			Drivers: []opa.Driver{
				{
					ID:         "123345",
					Age:        23,
					Experience: 3,
				},
			},
			Vehicles: vehicles,
		},
	}, nil

}

func opaValidateDecision(data []byte) (opa.OPARequest, error) {
	var req Request

	if err := json.Unmarshal(data, &req); err != nil {
		return opa.OPARequest{}, err
	}

	var vehicles []opa.Vehicle

	for _, v := range req.Vehicles {
		vehicles = append(vehicles, opa.Vehicle{
			ID:       v.VIN,
			BodyType: v.BodyType,
			Class:    v.Class,
			Amount:   v.Amount,
		})
	}

	return opa.OPARequest{
		Input: opa.Input{
			Operation:   req.Operation,
			Commodities: req.Commodities,
			Drivers: []opa.Driver{
				{
					ID:         "123345",
					Age:        23,
					Experience: 3,
				},
			},
			Vehicles: vehicles,
		},
	}, nil

}

func main() {
	ctx := context.Background()

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(7070),
	)

	s.RegisterSubRouter("/api/v1", getRoutes(s.Logger), middleware.RequestID)

	errChan := make(chan error, 1)
	go s.Serve(errChan)
	s.AutoHandleErrors(ctx, errChan)
}
