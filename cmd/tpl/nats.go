// Copyright 2025 Sencillo
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
	return []byte(`{{ $tick := "` + "`" + `" -}}
package service

import (
	"encoding/json"
	"time"

	cwerrors "github.com/CoverWhale/coverwhale-go/errors"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type MathRequest struct {
	A int {{ $tick }}json:"a"{{ $tick }}
	B int {{ $tick }}json:"b"{{ $tick }}
}

type MathResponse struct {
	Result int {{ $tick }}json:"result"{{ $tick }}
}

func SpecificHandler(logger *logr.Logger, r micro.Request) error {
	r.Respond([]byte("in the specific handler"))

	return nil
}

func Add(logger *logr.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwerrors.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A + mr.B}

	r.RespondJSON(resp)

	return nil
}

func Subtract(logger *logr.Logger, r micro.Request) error {
	var mr MathRequest
	if err := json.Unmarshal(r.Data(), &mr); err != nil {
		return cwerrors.NewClientError(err, 400)
	}

	resp := MathResponse{Result: mr.A - mr.B}

	r.RespondJSON(resp)
	return nil
}

func WatchForConfig(logger *logr.Logger, js nats.JetStreamContext) {
	kv, err := js.KeyValue("configs")
	if err != nil {
		logr.Fatal(err)
	}

	w, err := kv.Watch("{{ .Name }}.log_level")
	if err != nil {
		logr.Fatal(err)
	}

	for val := range w.Updates() {
		if val == nil {
			continue
		}

		level := string(val.Value())
		if level == "info" {
			logger.Level = logr.InfoLevel
		}

		if level == "error" {
			logger.Level = logr.ErrorLevel
		}

		if level == "debug" {
			logger.Level = logr.DebugLevel
		}

		logger.Infof("set log level to %s", level)
	}

	time.Sleep(5 * time.Second)
}
`)
}
