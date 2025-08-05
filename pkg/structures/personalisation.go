package structures

type ItemObject struct {
	ID             uint    `json:"id"`
	CafeID         uint    `json:"cafe_id"`
	ImageURL       string  `json:"image_url"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
	IsCustomizable bool    `json:"is_customizable"`
}

type PersonalisedDataResponse struct {
	ResponseKey string       `json:"response_key"`
	Items       []ItemObject `json:"items"`
}
