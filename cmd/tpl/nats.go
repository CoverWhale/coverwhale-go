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

package tpl

func Nats() []byte {
	return []byte(`package server
import (
  "context"
	"fmt"
	"log"

  cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
  "github.com/nats-io/nats.go/micro"
)

func handleRequest(req micro.Request) {
	logr.Infof("received request on %s", req.Subject())

	// you can return errors like this
	// if err != nil {
	//  req.Error("400", err.String(), nil)
	// }
	// return

	ctx := context.Background()
	doMore(ctx)

	response := fmt.Sprintf("%s yourself", string(req.Data()))

	req.Respond([]byte(response))
}

func NewMicro(conn *nats.Conn) (micro.Service, error) {
	config := micro.Config{
		Name:        "{{ .Name }}",
		Version:     "0.0.1",
		Description: "{{ .Name }}'s description",
		Endpoint: &micro.EndpointConfig{
			Subject: "prime.{{ .Name }}.doit",
			Handler: micro.HandlerFunc(handleRequest),
		},
	}

	svc, err := micro.AddService(conn, config)
	if err != nil {
		return svc, err
	}

	group := svc.AddGroup("prime.{{ .Name }}")
	if err := group.AddEndpoint("doit", micro.HandlerFunc(handleRequest)); err != nil {
		return svc, err
	}

	return svc, nil
}

func Watch(n *cwnats.NATSClient, s string) {
	logr.Infof("watching for requests on %s", s)
	_, err := n.Conn.Subscribe(s, HandleMessage)
	if err != nil {
		log.Printf("Error in subscribing: %v", err)
	}
}

func HandleMessage(m *nats.Msg) {
	logr.Infof("recevied request on %s", m.Subject)

	switch m.Subject {
	case "prime.{{ .Name }}.pub":
        fmt.Printf("received pub %s\n", string(m.Data))
	}
}
`)
}
