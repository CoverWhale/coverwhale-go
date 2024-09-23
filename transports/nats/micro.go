package nats

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cwerrors "github.com/CoverWhale/coverwhale-go/errors"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go/micro"
	"github.com/segmentio/ksuid"
)

type HandlerWithErrors func(*logr.Logger, micro.Request) error

type ClientError interface {
	Error() string
	Code() int
	Body() []byte
	Internal() string
}

func HandleNotify(s micro.Service, healthFuncs ...func(chan<- string, micro.Service)) error {
	stopChan := make(chan string, 1)
	for _, v := range healthFuncs {
		go v(stopChan, s)
	}

	go handleNotify(stopChan)

	logr.Info(<-stopChan)
	return s.Stop()
}

func handleNotify(stopChan chan<- string) {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	stopChan <- fmt.Sprintf("received signal: %v", sig)
}

// ErrorHandler wraps a normal micro endpoint and allows for returning errors natively. Errors are
// checked and if an error is a client error, details are returned, otherwise a 500 is returned and logged
func ErrorHandler(logger *logr.Logger, h HandlerWithErrors) micro.HandlerFunc {
	return func(r micro.Request) {
		start := time.Now()
		id, err := SubjectToRequestID(r.Subject())
		if err != nil {
			handleRequestError(logger, cwerrors.NewClientError(err, 400), r)
			return
		}
		reqLogger := logger.WithContext(map[string]string{"request_id": id, "path": r.Subject()})
		defer func() {
			reqLogger.Infof("duration %dms", time.Since(start).Milliseconds())
		}()

		err = h(reqLogger, r)
		if err == nil {
			return
		}

		handleRequestError(reqLogger, err, r)
	}
}

func handleRequestError(logger *logr.Logger, err error, r micro.Request) {
	ce, ok := err.(ClientError)
	if ok {
		logger.Error(ce.Internal())
		r.Error(fmt.Sprintf("%d", ce.Code()), http.StatusText(ce.Code()), ce.Body())
	}

	logger.Error(err)

	r.Error("500", "internal server error", []byte(`{"error": "internal server error"}`))
}

func SubjectToRequestID(s string) (string, error) {
	split := strings.Split(s, ".")
	if len(split) < 3 {
		return "", fmt.Errorf("invalid subject")
	}

	id := split[3]

	_, err := ksuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid ksuid request ID")
	}

	return id, nil
}

func RequestLogger(l *logr.Logger, subject string) (*logr.Logger, error) {
	id, err := SubjectToRequestID(subject)
	if err != nil {
		return nil, err
	}
	return l.WithContext(map[string]string{"request_id": id}), nil
}
