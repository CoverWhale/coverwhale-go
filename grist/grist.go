package grist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClientOpt func(*Client)

type Client struct {
	// Grist API token
	token string
	// Grist Server URL
	URL string
	// HTTP Client
	client *http.Client
}

type Request struct {
	Path     string
	Method   string
	Document string
	Table    string
	Filter   json.RawMessage
	Data     io.Reader
}

type RequestOpt func(*Request)

func NewClient(opts ...ClientOpt) *Client {
	c := &Client{}

	for _, v := range opts {
		v(c)
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	return c
}

func SetAPIKey(key string) ClientOpt {
	return func(c *Client) {
		c.token = key
	}
}

func SetURL(url string) ClientOpt {
	return func(c *Client) {
		c.URL = url
	}
}

func SetHTTPClient(h *http.Client) ClientOpt {
	return func(c *Client) {
		c.client = h
	}
}

func (c *Client) GetDocument(document string) (json.RawMessage, error) {
	request := Request{
		Path:   fmt.Sprintf("/docs/%s", document),
		Method: http.MethodGet,
	}
	return c.httpRequest(request)
}

func SetDocument(d string) RequestOpt {
	return func(r *Request) {
		r.Document = d
	}
}

func SetTable(t string) RequestOpt {
	return func(r *Request) {
		r.Table = t
	}
}

func SetFilter(f json.RawMessage) RequestOpt {
	return func(r *Request) {
		r.Filter = f
	}
}

func (c *Client) GetRecordsWithOptions(opts ...RequestOpt) (json.RawMessage, error) {
	var r Request

	for _, opt := range opts {
		opt(&r)
	}

	r.Path = fmt.Sprintf("/api/docs/%s/tables/%s/records", r.Document, r.Table)

	return c.getRecords(r)
}

func (c *Client) GetRecords(document, table string) (json.RawMessage, error) {

	request := Request{
		Path:   fmt.Sprintf("/api/docs/%s/tables/%s/records", document, table),
		Method: http.MethodGet,
	}

	return c.getRecords(request)
}

func (c *Client) GetFilteredRecords(document, table string, filter json.RawMessage) (json.RawMessage, error) {
	request := Request{
		Path:   fmt.Sprintf("/api/docs/%s/tables/%s/records", document, table),
		Method: http.MethodGet,
	}

	if filter != nil {
		request.Filter = filter
	}

	return c.getRecords(request)

}

func (c *Client) getRecords(r Request) (json.RawMessage, error) {
	return c.httpRequest(r)
}

func (c *Client) CreateRecord(document, table string, r io.Reader) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/docs/%s/tables/%s/records", c.URL, document)
	request := Request{
		Path:   path,
		Method: http.MethodPost,
		Data:   r,
	}
	return c.httpRequest(request)
}

func (c *Client) httpRequest(request Request) (json.RawMessage, error) {
	url := fmt.Sprintf("%s%s", c.URL, request.Path)
	token := fmt.Sprintf("Bearer %s", c.token)

	req, err := http.NewRequest(request.Method, url, request.Data)
	if err != nil {
		return nil, err
	}

	if request.Filter != nil {
		q := req.URL.Query()
		q.Add("filter", string(request.Filter))
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Add("Authorization", token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", http.StatusText(resp.StatusCode))
	}

	return body, nil
}
