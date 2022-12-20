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
	"fmt"
	"log"

    cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
)

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
	case "prime.{{ .Name }}.req":
		if err := HandleRequest(m); err != nil {
			logr.Errorf("error sending request: %v", err)
		}
    default:
        fmt.Println(string(m.Data))
	}
}

func HandleRequest(m *nats.Msg) error {
	data := fmt.Sprintf("%s yourself", string(m.Data))
	msg := &nats.Msg{
		Data: []byte(data),
	}
	if err := m.RespondMsg(msg); err != nil {
		return err
	}

	return nil
}
`)
}
