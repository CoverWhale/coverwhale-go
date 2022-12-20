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

package opa

const (
	SideCarOPA OPAURL = "http://localhost:8181"
	CentralOPA OPAURL = "http://opa.svc.cluster.local:8181"
)

type OPAURL string

type OPAResponse struct {
	Result Result `json:"result"`
}
type OPARequest struct {
	Input Input `json:"input"`
}

type Result struct {
	Allow bool     `json:"allow"`
	Deny  []string `json:"deny,omitempty"`
}

// Decision should be called at a /decision endpoint.
type Decision struct {
	Allowed  bool     `json:"allowed"`
	Denials  []string `json:"denials,omitempty"`
	Carriers []string `json:"carriers,omitempty"`
}

type Input struct {
	State       string    `json:"state"`
	Operation   string    `json:"operation"`
	Commodities []string  `json:"commodities"`
	Drivers     []Driver  `json:"drivers"`
	Vehicles    []Vehicle `json:"vehicles"`
	Trailers    []Trailer `json:"trailers"`
}

type Driver struct {
	ID         string   `json:"id"`
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
