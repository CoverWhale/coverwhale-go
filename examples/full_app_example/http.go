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
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	cwhttp "github.com/SencilloDev/sencillo-go/transports/http"
)

type clientHandlerFunc func(http.ResponseWriter, *http.Request, ClientManager) error

func getErrorDetails(err error) (int, string) {
	clientError, ok := err.(*cwhttp.ClientError)
	if !ok {
		log.Printf("An error ocurred: %v", err)
		return 500, http.StatusText(http.StatusInternalServerError)
	}

	return clientError.Status, clientError.Details
}

func clientHandler(h clientHandlerFunc, cm ClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r, cm)
		if err == nil {
			return
		}

		status, body := getErrorDetails(err)

		apiErrDetails := fmt.Sprintf(`{"error": "%s"}`, body)

		w.WriteHeader(status)
		w.Write([]byte(apiErrDetails))
	}
}

func (a *Application) createProduct(w http.ResponseWriter, r *http.Request) error {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return err
	}

	new := NewProduct().SetName(p.Name).SetDescription(p.Description).SetPrice(p.Price)

	new.Save(a.ProductManager)

	if err := json.NewEncoder(w).Encode(new); err != nil {
		return err
	}

	return nil
}

func getProductByID(w http.ResponseWriter, r *http.Request, pm ProductManager) {
	id := r.PathValue("id")

	p := GetProduct(id, pm)

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Println(err)
	}
}

func getProducts(w http.ResponseWriter, r *http.Request, pm ProductManager) {
	p := GetAllProducts(pm)

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Println(err)
	}
}

func createClient(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	var c Client

	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return cwhttp.NewClientError(err, http.StatusBadRequest)
	}

	a, err := NewClient(c.Name, SetClientProducts(c.Products))
	if err != nil {
		return cwhttp.NewClientError(err, http.StatusBadRequest)
	}

	if err := a.Save(cm); err != nil {
		return err
	}

	if err := json.NewEncoder(w).Encode(a); err != nil {
		return err
	}

	return nil
}

func getClients(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	clients := GetAllClients(cm)

	if err := json.NewEncoder(w).Encode(clients); err != nil {
		return err
	}

	return nil
}

func getClientByID(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	id := r.PathValue("id")
	client := GetClient(id, cm)

	if err := json.NewEncoder(w).Encode(client); err != nil {
		return err
	}

	return nil
}

func (a *Application) buildRoutes(l *slog.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodGet,
			Path:    "/products/{id}",
			Handler: cwhttp.HandleWithContext(getProductByID, a.ProductManager),
		},
		{
			Method:  http.MethodGet,
			Path:    "/products",
			Handler: cwhttp.HandleWithContext(getProducts, a.ProductManager),
		},
		{
			Method:  http.MethodPost,
			Path:    "/clients",
			Handler: clientHandler(createClient, a.ClientManager),
		},
		{
			Method:  http.MethodGet,
			Path:    "/clients",
			Handler: clientHandler(getClients, a.ClientManager),
		},
		{
			Method:  http.MethodGet,
			Path:    "/clients/{id}",
			Handler: clientHandler(getClientByID, a.ClientManager),
		},
		{
			Method: http.MethodPost,
			Path:   "/products",
			Handler: &cwhttp.ErrHandler{
				Handler: a.createProduct,
				Logger:  l,
			},
		},
	}
}
