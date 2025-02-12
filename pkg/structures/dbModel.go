package structures

import (
	"time"

	"gorm.io/datatypes"
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
}

type Preference struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	PreferenceType  string    `gorm:"type:varchar(50);not null" json:"preference_type"`
	PreferenceValue string    `gorm:"type:varchar(100);not null" json:"preference_value"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
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
}

type ItemCustomization struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID        uint      `gorm:"not null" json:"menu_item_id"`
	CustomizationType string    `gorm:"type:varchar(50);not null" json:"customization_type"`
	OptionName        string    `gorm:"type:varchar(50);not null" json:"option_name"`
	AdditionalCost    float64   `gorm:"type:decimal(10,2);default:0" json:"additional_cost"`
	Priority          int       `gorm:"default:1" json:"priority"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CrossSell struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BaseItemID        uint      `gorm:"not null" json:"base_item_id"`
	CrossSellItemID   uint      `gorm:"not null" json:"cross_sell_item_id"`
	CrossSellCategory string    `gorm:"type:varchar(100)" json:"cross_sell_category"`
	Priority          int       `gorm:"default:1" json:"priority"`
	Description       string    `gorm:"type:text" json:"description"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
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

type CuratedCart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CafeID    uint           `gorm:"index;not null" json:"cafe_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	TimeOfDay TimeOfDay      `gorm:"type:varchar(50);not null" json:"time_of_day"`
	Date      time.Time      `gorm:"type:date;index" json:"date"`
	Source    string         `gorm:"type:varchar(50);default:'ai'" json:"source"`
	ItemIDs   datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
}
type CuratedCartItem struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CartID   uint `gorm:"index;not null" json:"cart_id"`
	ItemID   uint `gorm:"index;not null" json:"item_id"`
	Priority int  `gorm:"default:0" json:"priority"` // Helps in ordering items within a cart
}
