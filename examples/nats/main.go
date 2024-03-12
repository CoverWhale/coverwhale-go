package main

import (
	"encoding/json"

	cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	options := cwnats.MicroOptions{
		Servers:     nats.DefaultURL,
		BaseSubject: "prime.example",
		Config: micro.Config{
			Name:        "example-app",
			Version:     "0.0.1",
			Description: "An example application",
			Endpoint: &micro.EndpointConfig{
				Subject: "prime.example.generic",
				Handler: micro.HandlerFunc(func(r micro.Request) { r.Respond([]byte("responding")) }),
			},
		},
	}

	ms, err := cwnats.NewMicroService(options)
	if err != nil {
		logr.Fatal(err)
	}

	// add a singular handler as an endpoint
	ms.Service.AddEndpoint("specific", cwnats.ErrorHandler(specificHandler), micro.WithEndpointSubject("prime.example.specific"))

	// add a handler group
	grp := ms.Service.AddGroup("prime.example.math")
	grp.AddEndpoint("add", cwnats.ErrorHandler(add))
	grp.AddEndpoint("subtract", cwnats.ErrorHandler(subtract))

	ms.HandleNotify()
}

func specificHandler(r micro.Request) error {
	r.Respond([]byte("in the specific handler"))

	return nil
}

type MathRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type MathResponse struct {
	Result int `json:"result"`
}

func add(r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwnats.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A + mr.B}

	r.RespondJSON(resp)

	return nil
}

func subtract(r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwnats.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A - mr.B}

	r.RespondJSON(resp)
	return nil
}
