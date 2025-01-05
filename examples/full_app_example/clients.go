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
	"fmt"
)

var (
	ErrNameRequired = fmt.Errorf("name and age must be supplied")
)

type ClientOption func(*Client)

type ClientManager interface {
	ClientGetter
	ClientUpdater
	ClientLister
}

type ClientGetter interface {
	GetClient(id string) *Client
}

type ClientUpdater interface {
	UpdateClient(*Client) error
}

type ClientLister interface {
	GetAllClients() []*Client
}

type Client struct {
	ID       string
	Name     string
	Products []string
}

func NewClient(name string, opts ...ClientOption) (*Client, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	client := &Client{
		ID:   NewID(),
		Name: name,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func SetClientProduct(id string) ClientOption {
	return func(c *Client) {
		c.Products = append(c.Products, id)
	}
}

func SetClientProducts(ids []string) ClientOption {
	return func(c *Client) {
		c.Products = append(c.Products, ids...)
	}
}

func (a *Client) Save(u ClientUpdater) error {
	return u.UpdateClient(a)
}

func GetClient(id string, getter ClientGetter) *Client {
	return getter.GetClient(id)
}

func GetAllClients(getter ClientLister) []*Client {
	return getter.GetAllClients()
}
