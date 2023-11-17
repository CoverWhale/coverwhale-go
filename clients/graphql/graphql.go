package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClientOption func(*GraphQLClient)

type GraphQLClient struct {
	client *http.Client
	URL    string
}

type gqlResponse struct {
	Errors []gqlError `json:"errors"`
}

type gqlError struct {
	Message string `json:"message"`
}

type Query struct {
	Query     string `json:"query"`
	Variables `json:"variables"`
}

type Variables struct {
	Data json.RawMessage `json:"data"`
}

func handleGraphQLErrors(b []byte) error {
	var g gqlResponse
	if err := json.Unmarshal(b, &g); err != nil {
		return err
	}

	if g.Errors == nil {
		return nil
	}

	var errors error
	for _, v := range g.Errors {
		errors = fmt.Errorf("%w", fmt.Errorf("%s", v.Message))
	}

	return errors
}

func NewGraphQLClient(url string, opts ...ClientOption) *GraphQLClient {
	c := &GraphQLClient{
		client: http.DefaultClient,
		URL:    url,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func SetHTTPClient(client *http.Client) ClientOption {
	return func(c *GraphQLClient) {
		c.client = client
	}
}

func (g *GraphQLClient) newPostRequest(url string, data []byte) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: %v", string(body))
	}

	if err := handleGraphQLErrors(body); err != nil {
		return nil, err
	}

	return body, nil
}

func (g *GraphQLClient) Query(query string, r io.Reader) ([]byte, error) {
	var b bytes.Buffer

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return g.query(query, b.Bytes())
}

func (g *GraphQLClient) query(query string, vars json.RawMessage) ([]byte, error) {
	q := Query{
		Query: query,
		Variables: Variables{
			Data: vars,
		},
	}

	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/query", g.URL)

	return g.newPostRequest(url, data)
}
