package structures

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone     string    `gorm:"type:varchar(15);not null" json:"phone"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Gender    string    `gorm:"type:varchar(100);not null" json:"gender"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Preference struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	PreferenceType  string    `gorm:"type:varchar(50);not null" json:"preference_type"`
	PreferenceValue string    `gorm:"type:varchar(100);not null" json:"preference_value"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type MenuItem struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeID         uint      `gorm:"not null" json:"cafe_id"`
	Category       string    `gorm:"type:varchar(50);not null" json:"category"`  // e.g., "Pizza", "Beverages"
	SubCategory    string    `gorm:"type:varchar(50)" json:"sub_category"`       // e.g., "Hot Coffee", "Non-Veg Pizza"
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`     // e.g., "Cappuccino"
	Description    string    `gorm:"type:text" json:"description"`               // Item description
	Price          float64   `gorm:"type:decimal(10,2);not null" json:"price"`   // Base price of the item
	IsCustomizable bool      `gorm:"default:false" json:"is_customizable"`       // Whether customization is allowed
	FoodType       string    `gorm:"type:varchar(10);not null" json:"food_type"` // e.g., "veg", "non-veg", "vegan"
	ImageURL       string    `gorm:"type:varchar(255)" json:"image_url"`         // URL for item image
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`           // Timestamp for item creation
}

type ItemCustomization struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID        uint      `gorm:"not null" json:"menu_item_id"`
	CustomizationType string    `gorm:"type:varchar(50);not null" json:"customization_type"`
	OptionName        string    `gorm:"type:varchar(50);not null" json:"option_name"`
	AdditionalCost    float64   `gorm:"type:decimal(10,2);default:0" json:"additional_cost"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Upsell struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BaseItemID   uint      `gorm:"not null" json:"base_item_id"`
	UpsellItemID uint      `gorm:"not null" json:"upsell_item_id"`
	UpsellType   string    `gorm:"type:varchar(50);not null" json:"upsell_type"`
	Priority     int       `gorm:"default:1" json:"priority"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CrossSell struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BaseItemID      uint      `gorm:"not null" json:"base_item_id"`
	CrossSellItemID uint      `gorm:"not null" json:"cross_sell_item_id"`
	Priority        int       `gorm:"default:1" json:"priority"`
	Description     string    `gorm:"type:text" json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Order struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      string    `gorm:"type:varchar(50);default:'Pending'" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type OrderItem struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        uint      `gorm:"not null" json:"order_id"`
	MenuItemID     uint      `gorm:"not null" json:"menu_item_id"`
	Quantity       int       `gorm:"not null" json:"quantity"`
	Price          float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Customizations string    `gorm:"type:jsonb" json:"customizations"` // Store customizations as JSON
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}
