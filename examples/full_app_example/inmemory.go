// Copyright 2023 Cover Whale Insurance Solutions Inc.
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

type InMemory struct {
	Clients  []*Client
	Products []*Product
}

func NewInMemoryStore() *InMemory {
	return &InMemory{}
}

func (i *InMemory) GetClient(id string) *Client {
	for _, v := range i.Clients {
		if v.ID == id {
			return v
		}
	}

	return nil

}

func (i *InMemory) UpdateClient(a *Client) error {
	i.Clients = append(i.Clients, a)

	return nil
}

func (i *InMemory) GetAllClients() []*Client {
	return i.Clients
}

func (i *InMemory) GetProduct(id string) *Product {
	for _, v := range i.Products {
		if v.ID == id {
			return v
		}
	}

	return nil
}

func (i *InMemory) UpdateProduct(p *Product) error {
	i.Products = append(i.Products, p)

	return nil
}

func (i *InMemory) SearchProducts(name string) []*Product {
	var products []*Product

	for _, v := range i.Products {
		if v.Name == name {
			products = append(products, v)
		}
	}

	return products
}

func (i *InMemory) GetAllProducts() []*Product {
	return i.Products
}
