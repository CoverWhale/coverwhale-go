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
