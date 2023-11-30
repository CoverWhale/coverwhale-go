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
