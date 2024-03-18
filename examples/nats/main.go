package main

import (
	"encoding/json"

	cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/invopop/jsonschema"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func schemaString(s any) string {
	schema := jsonschema.Reflect(s)
	data, err := schema.MarshalJSON()
	if err != nil {
		logr.Fatal(err)
	}

	return string(data)
}

func main() {

	logger := logr.NewLogger()
	config := micro.Config{
		Name:        "example-app",
		Version:     "0.0.1",
		Description: "An example application",
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		logr.Fatal(err)
	}
	defer nc.Close()

	svc, err := micro.AddService(nc, config)
	if err != nil {
		logr.Fatal(err)
	}

	// add a singular handler as an endpoint
	svc.AddEndpoint("specific", cwnats.ErrorHandler(logger, specificHandler), micro.WithEndpointSubject("prime.example.specific"))

	// add a handler group
	grp := svc.AddGroup("prime.example.math", micro.WithGroupQueueGroup("example"))
	grp.AddEndpoint("add",
		cwnats.ErrorHandler(logger, add),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "adds two numbers",
			"format":          "application/json",
			"request_schema":  schemaString(&MathRequest{}),
			"response_schema": schemaString(&MathResponse{}),
		}))
	grp.AddEndpoint("subtract",
		cwnats.ErrorHandler(logger, subtract),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "subtracts two numbers",
			"format":          "application/json",
			"request_schema":  schemaString(&MathRequest{}),
			"response_schema": schemaString(&MathResponse{}),
		}))

	cwnats.HandleNotify(svc)
}

func specificHandler(logger *logr.Logger, r micro.Request) error {
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

func add(logger *logr.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwnats.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A + mr.B}

	r.RespondJSON(resp)

	return nil
}

func subtract(logger *logr.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwnats.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A - mr.B}

	r.RespondJSON(resp)
	return nil
}
