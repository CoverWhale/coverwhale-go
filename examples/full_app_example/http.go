package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/go-chi/chi/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type clientHandlerFunc func(http.ResponseWriter, *http.Request, ClientManager) error

func clientHandler(h clientHandlerFunc, cm ClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r, cm)

		if err == nil {
			return
		}

		log.Println(err)
		return
	}
}

func (a *Application) createProduct(w http.ResponseWriter, r *http.Request) error {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return err
	}

	new := NewProduct().SetName(p.Name).SetDescription(p.Description).SetPrice(p.Price)

	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "products",
		Operation:  "create",
	}

	new.Save(a.ProductManager)
	s.End()

	if err := json.NewEncoder(w).Encode(new); err != nil {
		return err
	}

	return nil
}

func getProductByID(w http.ResponseWriter, r *http.Request, pm ProductManager) {
	id := chi.URLParam(r, "id")
	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "products",
		Operation:  "get by id",
		QueryParameters: map[string]interface{}{
			"id": id,
		},
	}

	p := GetProduct(id, pm)
	s.End()

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Println(err)
	}
}

func getProducts(w http.ResponseWriter, r *http.Request, pm ProductManager) {
	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "products",
		Operation:  "get all",
	}
	p := GetAllProducts(pm)
	s.End()

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Println(err)
	}
}

func createClient(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	var c Client

	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return err
	}

	a, err := NewClient(c.Name, SetClientProducts(c.Products))
	if err != nil {
		return err
	}

	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "clients",
		Operation:  "create",
	}
	a.Save(cm)
	s.End()

	if err := json.NewEncoder(w).Encode(a); err != nil {
		return err
	}

	return nil
}

func getClients(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "clients",
		Operation:  "get all",
	}
	clients := GetAllClients(cm)
	s.End()

	if err := json.NewEncoder(w).Encode(clients); err != nil {
		return err
	}

	return nil
}

func getClientByID(w http.ResponseWriter, r *http.Request, cm ClientManager) error {
	id := chi.URLParam(r, "id")
	txn := newrelic.FromContext(r.Context())
	s := newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(txn),
		Product:    "in-memory",
		Collection: "clients",
		Operation:  "get by ID",
		QueryParameters: map[string]interface{}{
			"id": id,
		},
	}
	client := GetClient(id, cm)
	s.End()

	if err := json.NewEncoder(w).Encode(client); err != nil {
		return err
	}

	return nil
}

func (a *Application) buildRoutes(l *logging.Logger) []cwhttp.Route {
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
