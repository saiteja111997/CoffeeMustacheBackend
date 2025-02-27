package structures

import "gorm.io/datatypes"

// Define request structure for multiple cart items
type AddToCartRequest struct {
	CartID      string            `json:"cart_id"`
	SessionID   string            `json:"session_id" validate:"required"`
	TotalAmount float64           `json:"total_amount" validate:"required"`
	Items       []CartItemRequest `json:"items" validate:"required"`
}

type CartItemRequest struct {
	CartItemId       string         `json:"cart_item_id"`
	ItemID           uint           `json:"item_id" validate:"required"`
	Quantity         int            `json:"quantity" validate:"required,min=1"`
	Price            float64        `json:"price" validate:"required"`
	AddedVia         string         `json:"added_via" validate:"required"`
	SpecialRequest   string         `json:"special_request"`
	CustomizationIDs datatypes.JSON `json:"customization_ids"`   // Expecting JSON array like ["1", "2", "3"]
	CrossSellItemIDs datatypes.JSON `json:"cross_sell_item_ids"` // Expecting JSON array like ["1", "2", "3"]
}

type GetCartRequest struct {
	CartID    string `json:"cart_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

type CartItemResponse struct {
	CartItemId           string              `json:"cart_item_id"`
	ItemID               uint                `json:"item_id"`
	Quantity             int                 `json:"quantity"`
	Price                float64             `json:"price"`
	AddedVia             string              `json:"added_via"`
	SpecialRequest       string              `json:"special_request"`
	CustomizationDetails []map[string]string `json:"customization_ids"`
	CrossSellItemIDs     []string            `json:"cross_sell_item_ids"`
}

type UpdateCustomizationsRequest struct {
	CartItemId       string   `json:"cart_item_id" validate:"required"`
	Price            float64  `json:"price" validate:"required"`
	CustomizationIDs []string `json:"customization_ids"` // Full replacement
	CartAmount       float64  `json:"cart_amount" validate:"required"`
}

type UpdateCrossSellItemsRequest struct {
	CartItemId       string   `json:"cart_item_id" validate:"required"`
	Price            float64  `json:"price" validate:"required"`
	CrossSellItemIDs []string `json:"cross_sell_item_ids"` // Full replacement
	CartAmount       float64  `json:"cart_amount" validate:"required"`
}

type UpdateQuantityRequest struct {
	CartItemId string  `json:"cart_item_id" validate:"required"`
	Quantity   int     `json:"quantity" validate:"required"`
	CartAmount float64 `json:"cart_amount" validate:"required"`
}
