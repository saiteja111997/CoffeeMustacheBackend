package structures

type CallWaiterRequest struct {
	SessionID   string `json:"session_id"`
	TableNumber string `json:"table_number"`
	RequestType string `json:"request_type"`
	CafeID      uint   `json:"cafe_id"`
}
