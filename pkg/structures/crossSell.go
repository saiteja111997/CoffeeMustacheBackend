package structures

type CrossSellRequest struct {
	ItemID int `json:"item_id"`
}

type SuggestedItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

type CrossSellResponse struct {
	Suggestions []SuggestedItem `json:"suggestions"`
}
