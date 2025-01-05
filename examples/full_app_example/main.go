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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cwhttp "github.com/SencilloDev/sencillo-go/transports/http"
)

func main() {
	ctx := context.Background()
	ds := NewInMemoryStore()
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))

	h := Application{
		ProductManager: ds,
		ClientManager:  ds,
	}

	cwServer := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	).RegisterSubRouter("/api/v1", h.buildRoutes(l))
	// Sencillo-go accepts chi middleware also
	// .RegisterSubRouter("/api/v1", h.buildRoutes(l), middleware.Logger, middleware.Throttle(1))

	h.Server = cwServer

	products(ds)

	errChan := make(chan error, 1)
	go h.Server.Serve(errChan)

	go func() {
		serverErr := <-errChan
		if serverErr != nil {
			cwServer.Logger.Error(fmt.Sprintf("error starting server: %v", serverErr))
			cwServer.ShutdownServer(ctx)
		}
	}()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	cwServer.Logger.Info(fmt.Sprintf("received signal: %s", sig))
	cwServer.ShutdownServer(ctx)

}
