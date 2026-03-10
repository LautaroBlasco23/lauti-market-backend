package domain

import (
	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
)

type Product struct {
	id          string
	storeID     string
	name        string
	description string
	category    string
	stock       int
	price       float64
	imageURL    *string
}

func NewProduct(id, storeID, name, description, category string, stock int, price float64, imageURL *string) (*Product, error) {
	if storeID == "" {
		return nil, apiDomain.ErrInvalidStoreID
	}
	if name == "" {
		return nil, apiDomain.ErrInvalidProductName
	}
	if description == "" {
		return nil, apiDomain.ErrInvalidProductDescription
	}
	if category == "" {
		return nil, apiDomain.ErrInvalidCategory
	}
	if stock < 0 {
		return nil, apiDomain.ErrInvalidStock
	}
	if price <= 0 {
		return nil, apiDomain.ErrInvalidPrice
	}

	return &Product{
		id:          id,
		storeID:     storeID,
		name:        name,
		description: description,
		category:    category,
		stock:       stock,
		price:       price,
		imageURL:    imageURL,
	}, nil
}

func (p *Product) ID() string          { return p.id }
func (p *Product) StoreID() string     { return p.storeID }
func (p *Product) Name() string        { return p.name }
func (p *Product) Description() string { return p.description }
func (p *Product) Category() string    { return p.category }
func (p *Product) Stock() int          { return p.stock }
func (p *Product) Price() float64      { return p.price }
func (p *Product) ImageURL() *string   { return p.imageURL }

func (p *Product) Update(name, description, category string, stock int, price float64, imageURL *string) error {
	if name == "" {
		return apiDomain.ErrInvalidProductName
	}
	if description == "" {
		return apiDomain.ErrInvalidProductDescription
	}
	if category == "" {
		return apiDomain.ErrInvalidCategory
	}
	if stock < 0 {
		return apiDomain.ErrInvalidStock
	}
	if price <= 0 {
		return apiDomain.ErrInvalidPrice
	}

	p.name = name
	p.description = description
	p.category = category
	p.stock = stock
	p.price = price
	p.imageURL = imageURL
	return nil
}
