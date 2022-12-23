package main

type ProductService struct {
	manager ProductManager
}

type ProductManager interface {
	ProductGetter
	ProductLister
	ProductUpdater
}

type ProductGetter interface {
	GetProduct(id string) *Product
}

type ProductLister interface {
	GetAllProducts() []*Product
}

type ProductUpdater interface {
	UpdateProduct(*Product) error
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	Description string  `json:"description"`
}

func NewProduct() *Product {
	return &Product{
		ID: NewID(),
	}
}

func (p *Product) SetName(n string) *Product {
	p.Name = n
	return p
}

func (p *Product) SetPrice(price float32) *Product {
	p.Price = price
	return p
}

func (p *Product) SetDescription(desc string) *Product {
	p.Description = desc
	return p
}

func (p *Product) Save(u ProductUpdater) error {
	return u.UpdateProduct(p)
}

func GetProduct(id string, pm ProductGetter) *Product {
	return pm.GetProduct(id)
}

func GetAllProducts(pm ProductLister) []*Product {
	return pm.GetAllProducts()
}

func products(ds ProductUpdater) {
	p := []struct {
		name  string
		desc  string
		price float32
	}{
		{name: "pencil", desc: "a pencil", price: .50},
		{name: "ipad", desc: "an apple tablet", price: 300.00},
		{name: "iphone", desc: "an apple phone", price: 1100.00},
	}

	for _, v := range p {
		product := NewProduct().SetName(v.name).SetDescription(v.desc).SetPrice(v.price)
		product.Save(ds)
	}

}
