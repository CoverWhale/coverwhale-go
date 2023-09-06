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
