package nats

import (
	"github.com/nats-io/nats.go"
)

type LogWriter struct {
	Subject string
	Conn    *nats.Conn
}

func (l LogWriter) Write(p []byte) (int, error) {
	err := l.Conn.Publish(l.Subject, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
