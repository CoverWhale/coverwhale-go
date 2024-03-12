package nats

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type NATSMicro struct {
	Conn    *nats.Conn
	Service micro.Service
	Logger  *logr.Logger
}

type MicroOptions struct {
	BaseSubject string
	Config      micro.Config
	Servers     string
	Opts        []nats.Option
}

type HandlerWithErrors func(r micro.Request) error

type ClientError struct {
	Code    int
	Details string
}

func (c *ClientError) Error() string {
	return c.Details
}

func (c *ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, c.Details))
}

func (c *ClientError) CodeString() string {
	return strconv.Itoa(c.Code)
}

func (c ClientError) As(target error) bool {
	_, ok := target.(*ClientError)
	return ok
}

func NewclientError(err error, code int) ClientError {
	return ClientError{
		Code:    code,
		Details: err.Error(),
	}
}

func healthz(r micro.Request) {
	data := []byte("ok")
	headers := map[string][]string{
		"Nats-Service-Status-Code": {"ok"},
	}

	r.Respond(data, micro.WithHeaders(headers))
}

func NewMicroService(options MicroOptions) (NATSMicro, error) {
	nc, err := nats.Connect(options.Servers, options.Opts...)
	if err != nil {
		return NATSMicro{}, err
	}

	svc, err := micro.AddService(nc, options.Config)
	if err != nil {
		return NATSMicro{}, err
	}

	svc.AddEndpoint("health", micro.HandlerFunc(healthz), micro.WithEndpointSubject(fmt.Sprintf("%s.healthz", options.BaseSubject)))

	return NATSMicro{
		Service: svc,
		Conn:    nc,
	}, nil
}

func (n *NATSMicro) HandleNotify() {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	logr.Infof("received signal: %s", sig)
	n.Service.Stop()
	n.Conn.Drain()
	n.Conn.Close()
}

// ErrorHandler wraps a normal micro endpoint and allows for returning errors natively. Errors are
// checked and if an error is a client error, details are returned, otherwise a 500 is returned and logged
func ErrorHandler(h HandlerWithErrors) micro.HandlerFunc {
	return func(r micro.Request) {
		err := h(r)
		if err == nil {
			return
		}

		var ce *ClientError
		if errors.As(err, &ce) {
			r.Error(ce.CodeString(), http.StatusText(ce.Code), ce.Body())
			return
		}

		logr.Error(err)

		r.Error("500", "internal server error", []byte(`{"error": "internal server error"}`))
	}
}
