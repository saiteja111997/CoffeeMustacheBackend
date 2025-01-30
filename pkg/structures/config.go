package structures

type Config struct {
	DB_USERNAME        string `json:"DB_USERNAME"`
	DB_PASSWORD        string `json:"DB_PASSWORD"`
	DB_HOSTNAME        string `json:"DB_HOSTNAME"`
	DB_PORT            string `json:"DB_PORT"`
	DATABASE           string `json:"DATABASE"`
	ORIGIN             string `json:"ORIGIN"`
	TWILIO_ACCOUNT_SID string `json:"TWILIO_ACCOUNT_SID"`
	TWILIO_AUTH_TOKEN  string `json:"TWILIO_AUTH_TOKEN"`
	TWILIO_SERVICES_ID string `json:"TWILIO_SERVICES_ID"`
	OPEN_AI_API_KEY    string `json:"OPEN_AI_API_KEY"`
}
