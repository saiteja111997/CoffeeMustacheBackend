package structures

// CustomizationCategory groups customizations under a category
type CustomizationCategory struct {
	Category string              `json:"category"`
	Items    []CustomizationItem `json:"items"`
}

// CustomizationItem represents an upsell option for a specific category
type CustomizationItem struct {
	ItemName       string  `json:"item_name"`
	AdditionalCost float64 `json:"additional_cost"`
}

// CustomizationResponse represents the API response format
type CustomizationResponse struct {
	ItemID         uint                    `json:"item_id"`
	Customizations []CustomizationCategory `json:"customizations"`
}
