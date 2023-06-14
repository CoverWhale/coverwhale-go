package tpl

func Nats() []byte {
	return []byte(`package server
import (
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/logging"
	"github.com/nats-io/nats.go"
)

type NatsBackend struct {
	Servers string
	Options []nats.Option
	Conn    *nats.Conn
	JS      nats.JetStreamContext
	Logger  *logging.Logger
}

func NewNatsBackend(s string, opts ...nats.Option) *NatsBackend {
	return &NatsBackend{
		Servers: s,
		Options: opts,
        Logger: logging.NewLogger(),
	}
}

func (n *NatsBackend) Connect() error {
	nc, err := nats.Connect(n.Servers, n.Options...)
	if err != nil {
		return err
	}

	n.Conn = nc
	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	n.JS = js

	return nil
}

func (n *NatsBackend) Watch(s string) {
	n.Logger.Infof("watching for requests on %s", s)
	_, err := n.Conn.Subscribe(s, n.HandleMessage)
	if err != nil {
		log.Printf("Error in subscribing: %v", err)
	}
}

func (n *NatsBackend) HandleMessage(m *nats.Msg) {
	n.Logger.Infof("recevied request on %s", m.Subject)

	switch m.Subject {
	case "test.pub":
        fmt.Printf("received pub %s\n", string(m.Data))
	case "test.req":
		if err := n.HandleRequest(m); err != nil {
			n.Logger.Errorf("error sending request: %v", err)
		}
    default:
        fmt.Println(string(m.Data))
	}
}

func (n *NatsBackend) HandleRequest(m *nats.Msg) error {
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
