package structures

import (
	"time"

	"gorm.io/datatypes"
)

// UserStatus Enum
type UserStatus string

const (
	UserActive   UserStatus = "Active"
	UserInactive UserStatus = "Inactive"
)

// SessionStatus Enum
type SessionStatus string

const (
	Active   SessionStatus = "Active"
	Inactive SessionStatus = "Inactive"
)

type CartInsertType string

const (
	Direct              CartInsertType = "Direct"
	FromCuratedCart     CartInsertType = "FromCuratedCart"
	CrossSellFocus      CartInsertType = "CrossSellFocus"
	TopPicks            CartInsertType = "TopPicks"
	UpgradeCartAi       CartInsertType = "UpgradeCartAi"
	CrossSellCheckout   CartInsertType = "CrossSellCheckout"
	AddedByWaiter       CartInsertType = "AddedByWaiter"
	UpgradeCartAiWaiter CartInsertType = "UpgradeCartAiWaiter"
)

type CartItemStatus string

const (
	CartItemActive    CartItemStatus = "Active"
	CartItemOrdered   CartItemStatus = "Ordered"
	CartItemCanceled  CartItemStatus = "Canceled"
	CartItemDelivered CartItemStatus = "Delivered"
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
	OrderCancelled OrderStatus = "Cancelled"
	OrderConfirmed OrderStatus = "Confirmed"
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
	Completed PaymentStatus = "paid"
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
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone         string    `gorm:"type:varchar(15);not null" json:"phone"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	Gender        string    `gorm:"type:varchar(100);not null" json:"gender"`
	TermsAccepted bool      `gorm:"default:false" json:"terms_accepted"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type TermsAndConditions struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Version    string    `gorm:"type:varchar(50);not null" json:"version"`
	AcceptedOn time.Time `gorm:"autoCreateTime" json:"accepted_on"`
	UserID     uint      `gorm:"not null" json:"user_id"`
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
	ID               uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeID           uint           `gorm:"not null" json:"cafe_id"`
	Category         string         `gorm:"type:varchar(50);not null" json:"category"`
	Name             string         `gorm:"type:varchar(100);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	ShortDescription string         `gorm:"type:varchar(255)" json:"short_description"`
	Price            float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	IsCustomizable   bool           `gorm:"default:false" json:"is_customizable"`
	FoodType         string         `gorm:"type:varchar(10);not null" json:"food_type"`
	Cuisine          Cuisine        `gorm:"type:varchar(50)" json:"cuisine"`        // Cuisine as enum
	DietaryLabels    DietaryLabel   `gorm:"type:varchar(50)" json:"dietary_labels"` // Dietary label as enum
	SpiceLevel       SpiceLevel     `gorm:"type:varchar(20)" json:"spice_level"`    // Spice level as enum
	CMCategory       CMCategory     `gorm:"type:varchar(50)" json:"cm_category"`    // CM Category as enum
	Ingredients      string         `gorm:"type:text" json:"ingredients"`
	Allergens        datatypes.JSON `gorm:"type:jsonb" json:"allergens"`
	ServingSize      string         `gorm:"type:varchar(50)" json:"serving_size"`
	Calories         int            `gorm:"type:int" json:"calories"`
	PreparationTime  int            `gorm:"type:int" json:"preparation_time"`
	Discount         float64        `gorm:"type: decimal(5,2); default:0.0" json:"discount"`
	DiscountSection  string         `gorm:"type:varchar(255)" json:"discount_section"`
	PopularityScore  float64        `gorm:"default:0.0" json:"popularity_score"`
	ImageURL         string         `gorm:"type:varchar(255)" json:"image_url"`
	VideoUrl         string         `gorm:"type:varchar(255)" json:"video_url"`
	AvailableFrom    string         `gorm:"type:varchar(255)" json:"available_from"`
	AvailableTill    string         `gorm:"type:varchar(255)" json:"available_till"`
	AvailableAllDay  string         `gorm:"type:varchar(255)" json:"available_all_day"`
	IsAvailable      bool           `gorm:"default:true" json:"is_available"`
	KitchenArea      string         `gorm:"type:varchar(255)" json:"kitchen_area"`
	Tag              datatypes.JSON `gorm:"type:jsonb" json:"tag"`
	AudioURL         string         `gorm:"type:varchar(255)" json:"audio_url"`
	Rating           float64        `gorm:"default:0.0;not null" json:"rating"`
	TotalRatings     int            `gorm:"default:0;not null" json:"total_ratings"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
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
	ID              uint           `gorm:"primaryKey" json:"id"`
	CafeID          uint           `gorm:"index;not null" json:"cafe_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	TimeOfDay       TimeOfDay      `gorm:"type:varchar(50);not null" json:"time_of_day"`
	Date            time.Time      `gorm:"type:date" json:"date"`
	CartTotal       float64        `gorm:"type:decimal(10,2)" json:"cart_total"`
	DiscountedTotal float64        `gorm:"type:decimal(10,2)" json:"discounted_total"`
	DiscountPercent float64        `gorm:"type:decimal(10,2)" json:"discount_percent"`
	Source          string         `gorm:"type:varchar(50);default:'ai'" json:"source"`
	ButtonActions   int            `gorm:"type:int;index" json:"button_actions"`
	ItemIDs         datatypes.JSON `gorm:"type:jsonb"`
	ImageURL        string         `gorm:"type:varchar(255)" json:"image_url"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

type CuratedCartItem struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CartID   uint `gorm:"index;not null" json:"cart_id"`
	ItemID   uint `gorm:"index;not null" json:"item_id"`
	Priority int  `gorm:"default:0" json:"priority"` // Helps in ordering items within a cart
}

// Sessions Table
type Session struct {
	SessionID       string        `gorm:"primaryKey;type:varchar(100)" json:"session_id"`
	TableName       string        `gorm:"type:varchar(100);not null" json:"table_name"`
	CafeID          uint          `gorm:"not null" json:"cafe_id"`
	SessionStatus   SessionStatus `gorm:"type:varchar(50);not null" json:"session_status"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         *time.Time    `json:"end_time,omitempty"`
	PaymentImageURL string        `gorm:"type:varchar(255)" json:"payment_image_url"`
	CreatedBy       uint          `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// User Sessions Table
type UserSession struct {
	UserSessionID string     `gorm:"type:varchar(100);primaryKey" json:"user_session_id"`
	SessionID     string     `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID        uint       `gorm:"not null" json:"user_id"`
	JoinedAt      time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	LeftAt        *time.Time `json:"left_at,omitempty"`
	Status        UserStatus `json:"status"`
	Role          UserRole   `gorm:"type:varchar(50);not null" json:"role"`
}

// Cart Table
type Cart struct {
	CartID           string     `gorm:"type:varchar(100);primaryKey" json:"cart_id"`
	SessionID        string     `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID           uint       `gorm:"not null" json:"user_id"`
	CafeId           uint       `gorm:"not null" json:"cafe_id"`
	CartStatus       CartStatus `gorm:"type:varchar(50);not null" json:"cart_status"`
	TotalAmount      float64    `gorm:"type:decimal(10,2)" json:"total_amount"`
	DiscountAmount   float64    `gorm:"type:decimal(10,2)" json:"discount_amount"`
	Note             string     `gorm:"type:text" json:"note"`
	ModifiedByWaiter string     `gorm:"type:varchar(20)" json:"modified_by_waiter"`
	WaiterID         uint       `gorm:"not null" json:"waiter_id"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type CartItem struct {
	CartItemID       string         `gorm:"type:varchar(100);primaryKey" json:"cart_item_id"`
	CartID           string         `gorm:"type:varchar(100);not null" json:"cart_id"`
	ItemID           uint           `gorm:"not null" json:"item_id"`
	CafeID           uint           `json:"cafe_id"` // Removed not null
	Quantity         int            `gorm:"not null" json:"quantity"`
	Price            float64        `gorm:"type:decimal(10,2)" json:"price"`
	AddedAt          time.Time      `gorm:"autoCreateTime" json:"added_at"`
	AddedVia         CartInsertType `gorm:"type:varchar(50)" json:"added_via"`
	SpecialRequest   string         `gorm:"type:text" json:"special_request"`
	Status           CartItemStatus `gorm:"type:varchar(50)" json:"status"`
	CustomizationIDs datatypes.JSON `gorm:"type:jsonb" json:"customization_ids"`   // Customization IDs as JSON array
	CrossSellItemIDs datatypes.JSON `gorm:"type:jsonb" json:"cross_sell_item_ids"` // Cross Sell Item IDs as JSON array
	ModifiedByWaiter string         `gorm:"type:varchar(20)" json:"modified_by_waiter"`
	WaiterID         uint           `gorm:"type:int" json:"waiter_id"`
	IsDelivered      bool           `gorm:"default:false" json:"is_delivered"`
	KOTStatus        bool           `gorm:"default:false" json:"kot_status"`
	DeliveredAt      *time.Time     `gorm:"type:time" json:"delivered_at"` // Changed to pointer for optional value
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// Orders Table
type Order struct {
	OrderID        string        `gorm:"type:varchar(100);primaryKey" json:"order_id"`
	CartID         string        `gorm:"type:varchar(100);not null" json:"cart_id"`
	CafeId         uint          `gorm:"not null" json:"cafe_id"`
	SessionID      string        `gorm:"type:varchar(100);not null" json:"session_id"`
	UserID         uint          `gorm:"not null" json:"user_id"`
	WaiterID       uint          `gorm:"type:int" json:"waiter_id"`
	SpecialRequest string        `gorm:"type:text" json:"special_request"`
	OrderStatus    OrderStatus   `gorm:"type:varchar(50);not null" json:"order_status"`
	PaymentStatus  PaymentStatus `gorm:"type:varchar(50);not null" json:"payment_status"`
	PaymentMode    string        `gorm:"type:varchar(50);not null" json:"payment_mode"`
	IsClicked      bool          `gorm:"default:false" json:"is_clicked"`
	TotalAmount    float64       `gorm:"type:decimal(10,2)" json:"total_amount"`
	OrderTime      time.Time     `gorm:"autoCreateTime" json:"order_time"`
	CompletedTime  *time.Time    `json:"completed_time,omitempty"`
}

type UpdateCartResult struct {
	ID                    uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CartID                string    `gorm:"type:varchar(100);not null" json:"cart_id"`
	SuggestedItemID       uint      `gorm:"not null" json:"suggested_item_id"`
	SuggestedItemName     string    `gorm:"type:varchar(255)" json:"suggested_item_name"`
	SuggestedItemCategory string    `gorm:"type:varchar(100)" json:"suggested_item_category"`
	SuggestedItemPrice    float64   `gorm:"type:decimal(10,2)" json:"suggested_item_price"`
	AIResponse            string    `gorm:"type:jsonb;not null" json:"ai_response"`
	UserReason            string    `gorm:"type:text" json:"user_reason"`
	ReferenceReason       string    `gorm:"type:text" json:"reference_reason"`
	DiscountedPrice       float64   `gorm:"type:decimal(5,2)" json:"discounted_price"`
	DiscountPercent       float64   `gorm:"type:decimal(5,2)" json:"discount_percent"`
	UserAction            string    `gorm:"type:varchar(50);default:'pending'" json:"user_action"`   // "added", "ignored", "pending"
	WaiterAction          string    `gorm:"type:varchar(50);default:'pending'" json:"waiter_action"` // "added", "pending"
	WaiterID              uint      `gorm:"type:int" json:"waiter_id"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Discount struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeID        uint      `gorm:"not null" json:"cafe_id"`
	DiscountType  string    `gorm:"type:varchar(50)" json:"discount_type"`
	DiscountValue float64   `gorm:"type:decimal(5,2)" json:"discount_value"`
	TotalCost     float64   `gorm:"type:decimal(10,2)" json:"total_cost"`
	OrderId       string    `gorm:"type:varchar(100)" json:"order_id"`
	UserId        uint      `gorm:"not null" json:"user_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type MenuAIRecords struct {
	PromptId     string    `gorm:"primaryKey;type:varchar(100)" json:"prompt_id"`
	UserId       uint      `gorm:"not null" json:"user_id"`
	CafeId       uint      `gorm:"not null" json:"cafe_id"`
	GeneratedSql string    `gorm:"type:varchar(255)" json:"generated_sql"`
	Answer       string    `gorm:"type:text" json:"answer"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Prompt       string    `gorm:"type:text" json:"prompt"`
	UserResponse bool      `gorm:"default:false" json:"user_response"`
}

type ItemFeedback struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"` // User who gave feedback
	ItemID    uint      `gorm:"not null" json:"item_id"` // Item being rated
	Rating    int       `gorm:"not null" json:"rating"`  // Rating (1-5)
	SessionID string    `gorm:"not null" json:"session_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CafeFeedback struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`      // User who gave feedback
	CafeID      uint      `gorm:"not null" json:"cafe_id"`      // Cafe being rated
	Rating      int       `gorm:"not null" json:"rating"`       // Cafe rating (1-5)
	WouldReturn int       `gorm:"not null" json:"would_return"` // Would the user return? (0->No, 1->Yes, 2->No input)
	Review      string    `gorm:"type:text" json:"review"`      // Optional text review
	SessionID   string    `gorm:"not null" json:"session_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Cafe struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeCode     string         `gorm:"type:varchar(255)" json:"cafe_code"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Address      string         `gorm:"type:varchar(255)" json:"address"`
	City         string         `gorm:"type:varchar(100)" json:"city"`
	State        string         `gorm:"type:varchar(100)" json:"state"`
	Country      string         `gorm:"type:varchar(100)" json:"country"`
	ZipCode      string         `gorm:"type:varchar(20)" json:"zip_code"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Email        datatypes.JSON `gorm:"type:jsonb" json:"email"`
	OpeningTime  time.Time      `gorm:"type:time" json:"opening_time"`
	ClosingTime  time.Time      `gorm:"type:time" json:"closing_time"`
	Rating       float64        `gorm:"default:0.0" json:"rating"`
	ImageURL     string         `gorm:"type:varchar(255)" json:"image_url"`
	TotalRatings uint           `gorm:"default:0" json:"total_ratings"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Cafe) TableName() string {
	return "cafes"
}

type Table struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	CafeID        uint      `gorm:"not null" json:"cafe_id"`
	Capacity      uint      `gorm:"not null" json:"capacity"`
	SeatingAreaID uint      `gorm:"not null" json:"seating_area_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CustomerRequest struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	CafeID      uint      `gorm:"not null" json:"cafe_id"`
	TableNumber string    `gorm:"not null" json:"table_number"`
	SessionID   string    `gorm:"not null" json:"session_id"`
	IsClicked   bool      `gorm:"default:false" json:"is_clicked"`
	RequestType string    `gorm:"type:varchar(50);not null" json:"request_type"`
	RequestedAt time.Time `gorm:"autoCreateTime" json:"requested_at"`
}

type CafeAdvertisement struct {
	AdvertisementID string    `gorm:"primaryKey;type:varchar(100)" json:"advertisement_id"`
	Title           string    `gorm:"type:varchar(255);not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	ImageURL        string    `gorm:"type:varchar(255)" json:"image_url"`
	VideoURL        string    `gorm:"type:varchar(255)" json:"video_url"`
	CafeID          uint      `gorm:"not null" json:"cafe_id"`
	AdStartTime     time.Time `gorm:"type:timestamp" json:"ad_start_time"`
	AdEndTime       time.Time `gorm:"type:timestamp" json:"ad_end_time"`
	EventStartTime  string    `gorm:"type:varchar(20)" json:"event_start_time"` // Time when the event is scheduled
	EventEndTime    string    `gorm:"type:varchar(20)" json:"event_end_time"`
	EventDate       string    `gorm:"type:varchar(20)" json:"event_date"`                          // Date when the event is scheduled
	AdType          string    `gorm:"type:varchar(50);not null" json:"ad_type"`                    // e.g. "offer", "event", "popup"
	AdStatus        string    `gorm:"type:varchar(20);not null;default:'active'" json:"ad_status"` // e.g. "active", "inactive"
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CafeAdvertisementClick struct {
	ClickID         uint      `gorm:"primaryKey;autoIncrement" json:"click_id"`
	AdvertisementID string    `gorm:"type:varchar(100);not null" json:"advertisement_id"`
	UserID          uint      `gorm:"not null" json:"user_id"`
	CafeID          uint      `gorm:"not null" json:"cafe_id"`
	ClickedAt       time.Time `gorm:"autoCreateTime" json:"clicked_at"`
	IsInterested    bool      `gorm:"default:false" json:"is_interested"`  // Indicates if the user is interested in the ad
	ClickedCancel   bool      `gorm:"default:false" json:"clicked_cancel"` // Indicates if the user clicked cancel on the ad
}
