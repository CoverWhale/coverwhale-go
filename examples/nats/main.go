package main

import (
	"encoding/json"
	"log/slog"
	"os"

	sderrors "github.com/SencilloDev/sencillo-go/errors"
	sdnats "github.com/SencilloDev/sencillo-go/transports/nats"
	"github.com/invopop/jsonschema"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func schemaString(s any) string {
	schema := jsonschema.Reflect(s)
	data, err := schema.MarshalJSON()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return string(data)
}

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := micro.Config{
		Name:        "example-app",
		Version:     "0.0.1",
		Description: "An example application",
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	svc, err := micro.AddService(nc, config)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// add a singular handler as an endpoint
	svc.AddEndpoint("specific", sdnats.ErrorHandler(logger, specificHandler), micro.WithEndpointSubject("prime.example.specific"))

	// add a handler group
	grp := svc.AddGroup("prime.services.example.*.math", micro.WithGroupQueueGroup("example"))
	grp.AddEndpoint("add",
		sdnats.ErrorHandler(logger, add),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "adds two numbers",
			"format":          "application/json",
			"request_schema":  schemaString(&MathRequest{}),
			"response_schema": schemaString(&MathResponse{}),
		}))
	grp.AddEndpoint("subtract",
		sdnats.ErrorHandler(logger, subtract),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "subtracts two numbers",
			"format":          "application/json",
			"request_schema":  schemaString(&MathRequest{}),
			"response_schema": schemaString(&MathResponse{}),
		}))

	sdnats.HandleNotify(svc)
}

func specificHandler(logger *slog.Logger, r micro.Request) error {
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

func add(logger *slog.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return sderrors.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A + mr.B}

	r.RespondJSON(resp)

	return nil
}

func subtract(logger *slog.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return sderrors.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A - mr.B}

	r.RespondJSON(resp)
	return nil
}
