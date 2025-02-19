package structures

// Response Structures
type UpsellCategory struct {
	Name            string  `json:"name"`
	AdditionalCost  float64 `json:"additional_cost"`
	CustomizationID uint    `json:"customization_id"`
}

type CrossSellCategory struct {
	Name              string  `json:"name"`
	Priority          int     `json:"priority"`
	ItemID            uint    `json:"item_id"`
	CrossSellCategory string  `json:"cross_sell_category"`
	CMCategory        string  `json:"cm_category"`
	Price             float64 `json:"price"`
	ImageURL          string  `json:"image_url"`
}

type CartItems struct {
	ItemID     uint   `json:"item_id"`
	CMCategory string `json:"cm_category"`
}

type Categories struct {
	CMCategory string `json:"cm_category"`
}

type Response struct {
	Result struct {
		Upsell    map[string][]UpsellCategory    `json:"upsell"`
		CrossSell map[string][]CrossSellCategory `json:"crosssell"`
	} `json:"result"`
}
