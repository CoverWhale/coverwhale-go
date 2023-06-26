package tpl

// Because of the yaml, this file is indented using Spaces
func Client() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
package graph

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
    Errors []gqlError {{ $tick }}json:"errors"{{ $tick }}
}
type gqlError struct {
    Message string {{ $tick }}json:"message"{{ $tick }}
}

type Query struct {
    Query     string {{ $tick }}json:"query"{{ $tick }}
    Variables {{ $tick }}json:"variables"{{ $tick }}
}

type Variables struct {
    Data json.RawMessage {{ $tick }}json:"data"{{ $tick }}
}

// create a new gql client
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

// in case you want to use a bespoke http client
func SetHTTPClient(client *http.Client) ClientOption {
    return func(c *GraphQLClient) {
        c.client = client
    }
}

// http POST call
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

// handle the errors
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

// GraphQL query
func (g *GraphQLClient) query(query string, vars ...json.RawMessage) ([]byte, error) {
    q := Query{
        Query: query,
    }

    for _, v := range vars {
        q.Variables = Variables{Data: v}
    }

    data, err := json.Marshal(q)
    if err != nil {
        return nil, err
    }

    return g.newPostRequest(g.URL, data)
}
`)
}

// template from gqlgen tutorial
func GQLGen() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
# Where are all the schema files located? globs are supported eg  src/**/*.graphqls
schema:
  - graph/*.graphqls

# Where should the generated server code go?
exec:
  filename: graph/generated.go
  package: graph

# Uncomment to enable federation
# federation:
#   filename: graph/federation.go
#   package: graph

# Where should any generated models go?
model:
  filename: graph/models_gen.go
  package: graph

# Where should the resolver implementations go?
resolver:
  layout: follow-schema
  dir: graph
  package: graph
  filename_template: "{name}.resolvers.go"
  # Optional: turn on to not generate template comments above resolvers
  # omit_template_comment: false

# Optional: turn on use ` + "`" + `gqlgen:"fieldName"` + "`" + ` tags in your models
# struct_tag: json

# Optional: turn on to use []Thing instead of []*Thing
# omit_slice_element_pointers: false

# Optional: turn on to skip generation of ComplexityRoot struct content and Complexity function
# omit_complexity: false

# Optional: turn on to not generate any file notice comments in generated files
# omit_gqlgen_file_notice: false

# Optional: turn on to exclude the gqlgen version in the generated file notice. No effect if {{ $tick }}omit_gqlgen_file_notice{{ $tick }} is true.
# omit_gqlgen_version_in_file_notice: false

# Optional: turn off to make struct-type struct fields not use pointers
# e.g. type Thing struct { FieldA OtherThing } instead of { FieldA *OtherThing }
# struct_fields_always_pointers: true

# Optional: turn off to make resolvers return values instead of pointers for structs
# resolvers_always_return_pointers: true

# Optional: turn on to return pointers instead of values in unmarshalInput
# return_pointers_in_unmarshalinput: false

# Optional: wrap nullable input fields with Omittable
# nullable_input_omittable: true

# Optional: set to speed up generation time by not performing a final validation pass.
# skip_validation: true

# Optional: set to skip running {{ $tick }}go mod tidy{{ $tick }} when generating server code
# skip_mod_tidy: true

# gqlgen will search for any type names in the schema in these go packages
# if they match it will use them, otherwise it will generate them.
autobind:
#  - "github.com/CoverWhale/{{ .Name }}/graph"

# This section declares type mapping between the GraphQL and go type systems
#
# The first line in each type will be used as defaults for resolver arguments and
# modelgen, the others will be allowed when binding to fields. Configure them to
# your liking
models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  Int:
    model:
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
`)
}

func SchemaGraphqls() []byte {
	return []byte(`# GraphQL schema example
# GraphQL schema example
#
# https://gqlgen.com/getting-started/

type Todo  {
    id: ID!
    text: String!
    done: Boolean!
    user: User!
}

type User {
    id: ID!
    name: String!
}

type Query {
    todos: [Todo!]!
}

input NewTodo {
    text: String!
    userId: String!
}

type Mutation {
    createTodo(input: NewTodo!): Todo!
}
`)
}

// needed to import gql
func Tools() []byte {
	return []byte(`
//go:build tools
// +build tools

package tools

import (
    _ "github.com/99designs/gqlgen"
    _ "github.com/99designs/gqlgen/graphql/introspection"
)
`)
}

func Resolvers() []byte {
	return []byte(`package graph
// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
    todos []*Todo
}
`)
}

// we need this for the package import in resolvers
func ModelsGen() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
package graph   
// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.


type NewTodo struct {
  Text   string {{ $tick }}json:"text"{{ $tick }}
  UserID string {{ $tick }}json:"userId"{{ $tick }}
}

type User struct {
  ID   string {{ $tick }}json:"id"{{ $tick }}
  Name string {{ $tick }}json:"name"{{ $tick }}
}

type Todo struct {
  ID   string {{ $tick }}json:"id"{{ $tick }}
  Text string {{ $tick }}json:"text"{{ $tick }}
  Done bool   {{ $tick }}json:"done"{{ $tick }}
  User *User  {{ $tick }}json:"user"{{ $tick }}
}


`)
}

func SchemaResolvers() []byte {
	return []byte(`package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.33

import (
    "context"
	"fmt"
	"math/big"
    "math/rand"

)

// CreateTodo is the resolver for the createTodo field.
func (r *mutationResolver) CreateTodo(ctx context.Context, input NewTodo) (*Todo, error) {
	rand, _ := rand.Intn(100)
	todo := &Todo{
		Text: input.Text,
		ID:   fmt.Sprintf("T%d", rand),
		User: &User{ID: input.UserID, Name: "user " + input.UserID},
	}
	r.todos = append(r.todos, todo)
	return todo, nil
}

// Todos is the resolver for the todos field.
func (r *queryResolver) Todos(ctx context.Context) ([]*Todo, error) {
	return r.todos, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

`)
}
