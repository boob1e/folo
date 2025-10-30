package ordering

import "gorm.io/gorm"

// Basket represents a shopping basket
type Basket struct {
	gorm.Model
	Description string       `gorm:"column:description;type:text" json:"description"`
	BasketItems []BasketItem `json:"basketItems"`
}

// BasketItem represents an item in a basket
type BasketItem struct {
	gorm.Model
	BasketID   uint `json:"-"`
	MenuItemID uint `json:"-"`
	MenuItem   MenuItem
	Quantity   int
}

// MenuItem represents a menu item that can be added to a basket
type MenuItem struct {
	gorm.Model
	SKU   int    `gorm:"column:sku;not null" json:"sku"`
	Name  string `gorm:"column:name;not null" json:"name"`
	Price int    `gorm:"column:price;not null" json:"price"` // Price in cents
}
