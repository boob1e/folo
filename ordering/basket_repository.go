package ordering

import (
	"context"

	"gorm.io/gorm"
)

// BasketRepository handles database operations for baskets
type BasketRepository interface {
	Create(basket *Basket) error
	FindByID(id uint) (*Basket, error)
	FindByIDWithItems(id uint) (*Basket, error)
	FindAll(limit int) ([]Basket, error)
	Update(basket *Basket) error
	Delete(id uint) error
}

type basketRepository struct {
	db *gorm.DB
}

// NewBasketRepository creates a new basket repository
func NewBasketRepository(db *gorm.DB) BasketRepository {
	return &basketRepository{db: db}
}

// Create creates a new basket in the database
func (r *basketRepository) Create(basket *Basket) error {
	return r.db.Create(basket).Error
}

// FindByID finds a basket by ID
func (r *basketRepository) FindByID(id uint) (*Basket, error) {
	ctx := context.Background()
	basket, err := gorm.G[Basket](r.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &basket, nil
}

// FindByIDWithItems finds a basket by ID with all items and menu details preloaded
func (r *basketRepository) FindByIDWithItems(id uint) (*Basket, error) {
	var basket Basket
	err := r.db.Preload("BasketItems.MenuItem").First(&basket, id).Error
	return &basket, err
}

// FindAll returns all baskets with a limit
func (r *basketRepository) FindAll(limit int) ([]Basket, error) {
	ctx := context.Background()
	baskets, err := gorm.G[Basket](r.db).
		Preload("BasketItems", nil).
		Preload("BasketItems.MenuItem", nil).
		Limit(limit).
		Find(ctx)
	return baskets, err
}

// Update updates an existing basket
func (r *basketRepository) Update(basket *Basket) error {
	return r.db.Save(basket).Error
}

// Delete soft deletes a basket
func (r *basketRepository) Delete(id uint) error {
	return r.db.Delete(&Basket{}, id).Error
}
