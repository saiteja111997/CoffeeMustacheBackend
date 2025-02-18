package structures

import (
	"time"

	"gorm.io/datatypes"
)

// SessionStatus Enum
type SessionStatus string

const (
	Active   SessionStatus = "Active"
	Inactive SessionStatus = "Inactive"
)

type CartInsertType string

const (
	Direct            CartInsertType = "Direct"
	CrossSellPopUp    CartInsertType = "CrossSellPopUp"
	FromCuratedCart   CartInsertType = "FromCuratedCart"
	CrossSellFocus    CartInsertType = "CrossSellFocus"
	TopPicks          CartInsertType = "TopPicks"
	UpgradeCartAi     CartInsertType = "UpgradeCartAi"
	CrossSellCheckout CartInsertType = "CrossSellCheckout"
)

type CartItemStatus string

const (
	CartItemActive   CartItemStatus = "Active"
	CartItemOrdered  CartItemStatus = "Ordered"
	CartItemCanceled CartItemStatus = "Canceled"
)

// Role Enum (User Session)
type UserRole string

const (
	Host  UserRole = "Host"
	Guest UserRole = "Guest"
)

// CartStatus Enum
type CartStatus string

const (
	CartActive  CartStatus = "Active"
	CartOrdered CartStatus = "Ordered"
)

// OrderStatus Enum
type OrderStatus string

const (
	OrderPlaced    OrderStatus = "Placed"
	OrderConfirmed OrderStatus = "Confirmed"
	OrderPreparing OrderStatus = "Preparing"
	OrderReady     OrderStatus = "Ready"
	OrderCompleted OrderStatus = "Completed"
	OrderCancelled OrderStatus = "Cancelled"
)

// PaymentMethod Enum
type PaymentMethod string

const (
	Cash   PaymentMethod = "Cash"
	Card   PaymentMethod = "Card"
	UPI    PaymentMethod = "UPI"
	Wallet PaymentMethod = "Wallet"
)

// PaymentStatus Enum
type PaymentStatus string

const (
	Pending   PaymentStatus = "Pending"
	Completed PaymentStatus = "Completed"
	Failed    PaymentStatus = "Failed"
)

// AvailabilityStatus Enum (Menu Item)
type AvailabilityStatus string

const (
	Available   AvailabilityStatus = "Available"
	Unavailable AvailabilityStatus = "Unavailable"
)

// Cuisine Enum
type Cuisine string

const (
	Italian       Cuisine = "italian"
	Mexican       Cuisine = "mexican"
	Indian        Cuisine = "indian"
	Chinese       Cuisine = "chinese"
	Japanese      Cuisine = "japanese"
	Mediterranean Cuisine = "mediterranean"
	Thai          Cuisine = "thai"
	French        Cuisine = "french"
	American      Cuisine = "american"
	Korean        Cuisine = "korean"
	Vietnamese    Cuisine = "vietnamese"
	MiddleEastern Cuisine = "middle-eastern"
	Greek         Cuisine = "greek"
	Spanish       Cuisine = "spanish"
)

// Dietary Labels Enum
type DietaryLabel string

const (
	GlutenFree       DietaryLabel = "gluten-free"
	HighProtein      DietaryLabel = "high-protein"
	Vegan            DietaryLabel = "vegan"
	Keto             DietaryLabel = "keto"
	LactoseFree      DietaryLabel = "lactose-free"
	LowCarb          DietaryLabel = "low-carb"
	LowFat           DietaryLabel = "low-fat"
	Organic          DietaryLabel = "organic"
	SugarFree        DietaryLabel = "sugar-free"
	Paleo            DietaryLabel = "paleo"
	Vegetarian       DietaryLabel = "vegetarian"
	Whole30          DietaryLabel = "whole30"
	DiabeticFriendly DietaryLabel = "diabetic-friendly"
)

// Spice Level Enum
type SpiceLevel string

const (
	Mild       SpiceLevel = "mild"
	Medium     SpiceLevel = "medium"
	Spicy      SpiceLevel = "spicy"
	ExtraSpicy SpiceLevel = "extra-spicy"
)

type TimeOfDay string

const (
	Morning   TimeOfDay = "morning"
	Afternoon TimeOfDay = "noon"
	Night     TimeOfDay = "night"
)

// Define the enum type
type CMCategory string

const (
	Beverages       CMCategory = "Beverages"
	BreakfastBrunch CMCategory = "Breakfast & Brunch"
	Appetizers      CMCategory = "Appetizers & Small Bites"
	MainCourse      CMCategory = "Main Course"
	BreadsSides     CMCategory = "Breads & Sides"
	Desserts        CMCategory = "Desserts & Sweets"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone     string    `gorm:"type:varchar(15);not null" json:"phone"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Gender    string    `gorm:"type:varchar(100);not null" json:"gender"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Preference struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	PreferenceType  string    `gorm:"type:varchar(50);not null" json:"preference_type"`
	PreferenceValue string    `gorm:"type:varchar(100);not null" json:"preference_value"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type MenuItem struct {
	ID              uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeID          uint         `gorm:"not null" json:"cafe_id"`
	Category        string       `gorm:"type:varchar(50);not null" json:"category"`
	SubCategory     string       `gorm:"type:varchar(50)" json:"sub_category"`
	Name            string       `gorm:"type:varchar(100);not null" json:"name"`
	Description     string       `gorm:"type:text" json:"description"`
	Price           float64      `gorm:"type:decimal(10,2);not null" json:"price"`
	IsCustomizable  bool         `gorm:"default:false" json:"is_customizable"`
	FoodType        string       `gorm:"type:varchar(10);not null" json:"food_type"`
	Cuisine         Cuisine      `gorm:"type:varchar(50)" json:"cuisine"`        // Cuisine as enum
	DietaryLabels   DietaryLabel `gorm:"type:varchar(50)" json:"dietary_labels"` // Dietary label as enum
	SpiceLevel      SpiceLevel   `gorm:"type:varchar(20)" json:"spice_level"`    // Spice level as enum
	CMCategory      CMCategory   `gorm:"type:varchar(50)" json:"cm_category"`    // CM Category as enum
	Ingredients     string       `gorm:"type:text" json:"ingredients"`
	Allergens       string       `gorm:"type:varchar(255)" json:"allergens"`
	ServingSize     string       `gorm:"type:varchar(50)" json:"serving_size"`
	Calories        int          `gorm:"type:int" json:"calories"`
	PreparationTime int          `gorm:"type:int" json:"preparation_time"`
	Discount        float64      `gorm:"type:decimal(5,2)" json:"discount"`
	PopularityScore float64      `gorm:"default:0.0" json:"popularity_score"`
	ImageURL        string       `gorm:"type:varchar(255)" json:"image_url"`
	AvailableFrom   string       `gorm:"type:varchar(255)" json:"available_from"`
	AvailableTill   string       `gorm:"type:varchar(255)" json:"available_till"`
	AvailableAllDay bool         `gorm:"default:true" json:"available_all_day"`
	IsAvailable     bool         `gorm:"default:true" json:"is_available"`
	Tag             string       `gorm:"type:varchar(255)" json:"tag"`
	Rating          float64      `gorm:"default:0.0;not null" json:"rating"`
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

type ItemCustomization struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID        uint      `gorm:"not null" json:"menu_item_id"`
	CustomizationType string    `gorm:"type:varchar(50);not null" json:"customization_type"`
	OptionName        string    `gorm:"type:varchar(50);not null" json:"option_name"`
	AdditionalCost    float64   `gorm:"type:decimal(10,2);default:0" json:"additional_cost"`
	Priority          int       `gorm:"default:1" json:"priority"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CrossSell struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BaseItemID        uint      `gorm:"not null" json:"base_item_id"`
	CrossSellItemID   uint      `gorm:"not null" json:"cross_sell_item_id"`
	CrossSellCategory string    `gorm:"type:varchar(100)" json:"cross_sell_category"`
	Priority          int       `gorm:"default:1" json:"priority"`
	Description       string    `gorm:"type:text" json:"description"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CuratedCart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CafeID    uint           `gorm:"index;not null" json:"cafe_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	TimeOfDay TimeOfDay      `gorm:"type:varchar(50);not null" json:"time_of_day"`
	Date      time.Time      `gorm:"type:date;index" json:"date"`
	Source    string         `gorm:"type:varchar(50);default:'ai'" json:"source"`
	ItemIDs   datatypes.JSON `gorm:"type:jsonb"`
	ImageURL  string         `gorm:"type:varchar(255)" json:"image_url"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

type CuratedCartItem struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CartID   uint `gorm:"index;not null" json:"cart_id"`
	ItemID   uint `gorm:"index;not null" json:"item_id"`
	Priority int  `gorm:"default:0" json:"priority"` // Helps in ordering items within a cart
}

// Sessions Table
type Session struct {
	SessionID     string        `gorm:"primaryKey;type:varchar(100)" json:"session_id"`
	TableID       string        `gorm:"type:varchar(100);not null" json:"table_id"`
	CafeID        uint          `gorm:"not null" json:"cafe_id"`
	SessionStatus SessionStatus `gorm:"type:varchar(50);not null" json:"session_status"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       *time.Time    `json:"end_time,omitempty"`
	CreatedBy     uint          `gorm:"not null" json:"created_by"`
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// User Sessions Table
type UserSession struct {
	UserSessionID string    `gorm:"type:varchar(100);primaryKey" json:"user_session_id"`
	SessionID     string    `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	JoinedAt      time.Time `gorm:"autoCreateTime" json:"joined_at"`
	Role          UserRole  `gorm:"type:varchar(50);not null" json:"role"`
}

// Cart Table
type Cart struct {
	CartID      string     `gorm:"type:varchar(100);primaryKey" json:"cart_id"`
	SessionID   string     `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	CartStatus  CartStatus `gorm:"type:varchar(50);not null" json:"cart_status"`
	TotalAmount float64    `gorm:"type:decimal(10,2)" json:"total_amount"`
	Note        string     `gorm:"type:text" json:"note"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type CartItem struct {
	CartItemID       uint           `gorm:"primaryKey;autoIncrement" json:"cart_item_id"`
	CartID           string         `gorm:"type:varchar(100);not null" json:"cart_id"`
	ItemID           uint           `gorm:"not null" json:"item_id"`
	Quantity         int            `gorm:"not null" json:"quantity"`
	Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
	AddedAt          time.Time      `gorm:"autoCreateTime" json:"added_at"`
	AddedVia         CartInsertType `gorm:"type:varchar(50)" json:"added_via"`
	SpecialRequest   string         `gorm:"type:text" json:"special_request"`
	Status           CartItemStatus `gorm:"type:varchar(50)" json:"status"`
	CustomizationIDs datatypes.JSON `gorm:"type:jsonb" json:"customization_ids"`   // Customization IDs as JSON array
	CrossSellItemIDs datatypes.JSON `gorm:"type:jsonb" json:"cross_sell_item_ids"` // Cross Sell Item IDs as JSON array
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// Orders Table
type Order struct {
	OrderID       string        `gorm:"type:varchar(100);primaryKey" json:"order_id"`
	CartID        string        `gorm:"type:varchar(100);not null" json:"cart_id"`
	SessionID     string        `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID        uint          `gorm:"not null" json:"user_id"`
	OrderStatus   OrderStatus   `gorm:"type:varchar(50);not null" json:"order_status"`
	TotalAmount   float64       `gorm:"type:decimal(10,2)" json:"total_amount"`
	PaymentMethod PaymentMethod `gorm:"type:varchar(50);not null" json:"payment_method"`
	PaymentStatus PaymentStatus `gorm:"type:varchar(50);not null" json:"payment_status"`
	OrderTime     time.Time     `gorm:"autoCreateTime" json:"order_time"`
	CompletedTime *time.Time    `json:"completed_time,omitempty"`
}

// Order Items Table
type OrderItem struct {
	OrderItemID      uint           `gorm:"primaryKey;autoIncrement" json:"order_item_id"`
	OrderID          string         `gorm:"type:varchar(100);not null" json:"order_id"`
	ItemID           uint           `gorm:"not null" json:"item_id"`
	Quantity         int            `gorm:"not null" json:"quantity"`
	Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
	SpecialRequest   string         `gorm:"type:text" json:"special_request"`
	CustomizationIDs datatypes.JSON `gorm:"type:jsonb" json:"customization_ids"` // Customization IDs as JSON array
}
