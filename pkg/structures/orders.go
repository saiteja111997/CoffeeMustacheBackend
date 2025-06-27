package structures

import "time"

// Request payload for demonstration
type FetchOrderDetailsRequest struct {
	SessionID string `json:"session_id"`
}

// Customization for demonstration
type Customization struct {
	ID         uint   `json:"id"`
	OptionName string `json:"option_name"`
}

type CrossSells struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// Final response shape: user -> list of OrderResponses
// or however you want to shape your final JSON
type UserOrdersMap map[uint][]OrderResponse

// CartItemDetail extends CartItem with parsed customizations
type CartItemDetail struct {
	ItemName       string          `json:"item_name"`
	CartItemID     string          `json:"cart_item_id"`
	ItemID         uint            `json:"item_id"`
	Quantity       int             `json:"quantity"`
	Price          float64         `json:"price"`
	SpecialRequest string          `json:"special_request"`
	Customizations []Customization `json:"customizations"`
	CrossSells     []CrossSells    `json:"cross_sells"`
}

// OrderResponse groups orders under a user
type OrderResponse struct {
	OrderID     string           `json:"order_id"`
	CartID      string           `json:"cart_id"`
	CartItems   []CartItemDetail `json:"cart_items"`
	Discount    float64          `json:"discount"`
	OrderedAt   time.Time        `json:"ordered_at"`
	TotalAmount float64          `json:"total_amount"`
}

// PlaceOrderRequest represents the request payload
type PlaceOrderRequest struct {
	CartID         string  `json:"cart_id"`
	CafeID         uint    `json:"cafe_id"`
	SessionID      string  `json:"session_id"`
	TotalAmount    float64 `json:"total_amount"`
	Discount       float64 `json:"discount"`
	SpecialRequest string  `json:"special_request"`
}

// PlaceOrderResponse represents the response payload
type PlaceOrderResponse struct {
	OrderID string `json:"order_id"`
	Rewards uint   `json:"rewards"` // Assuming rewards is a uint, adjust as necessary
}

// Build final
// For each user ID in results, we know their name. We can either do a second pass
// or rely on the fact that we stored userName in each OrderResponse.
type FinalResponse struct {
	UserID               uint            `json:"user_id"`
	CumilativeOrderTotal float64         `json:"cumilative_order_total"`
	Discount             float64         `json:"discount"`
	Orders               []OrderResponse `json:"orders"`
}
