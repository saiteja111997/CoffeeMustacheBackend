package structures

// Response Structures
type UpsellCategory struct {
	Name            string  `json:"name"`
	AdditionalCost  float64 `json:"additional_cost"`
	CustomizationID uint    `json:"customization_id"`
}

type CrossSellCategory struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	ItemID   uint   `json:"item_id"`
}

type Response struct {
	Result struct {
		Upsell    map[string][]UpsellCategory    `json:"upsell"`
		CrossSell map[string][]CrossSellCategory `json:"crosssell"`
	} `json:"result"`
}
