package main

import (
	"bytes"
	helper "coffeeMustacheBackend/pkg/helper"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// MenuItem represents a menu item in the database.
type MenuItem struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null"`
}

// CrossSell represents the cross-sell recommendation table.
type CrossSell struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	BaseItemID      uint      `gorm:"not null"`
	CrossSellItemID uint      `gorm:"not null"`
	Priority        int       `gorm:"default:1"`
	Description     string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

// SuggestedItem represents the AI's recommended items.
type SuggestedItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Priority    int    `json:"priority"`
	Description string `json:"description"`
}

// Database instance
var db *gorm.DB

// OpenAI API Key
var apiKey string

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading environment variables file")
	}

	apiKey = os.Getenv("OPEN_AI_API_KEY")
}

func main() {

	// Database connection
	DB_USERNAME := os.Getenv("DB_USERNAME")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOSTNAME := os.Getenv("DB_HOSTNAME")
	DB_PORT := os.Getenv("DB_PORT")
	DATABASE := os.Getenv("DATABASE")

	//ctx := context.Background()
	if err := waitForHost(DB_HOSTNAME, DB_PORT); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Connection established")

	db, err := helper.Open(helper.Config{
		Username: DB_USERNAME,
		Password: DB_PASSWORD,
		Hostname: DB_HOSTNAME,
		Port:     DB_PORT,
		Database: DATABASE,
	})

	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	// Fetch all menu items
	var menuItems []MenuItem
	if err := db.Find(&menuItems).Error; err != nil {
		log.Fatalf("Error fetching menu items: %v", err)
	}

	fmt.Printf("Processing %d menu items for cross-sell suggestions...\n", len(menuItems))

	// Process each menu item
	for _, item := range menuItems {
		fmt.Printf("Processing Item ID: %d - %s\n", item.ID, item.Name)

		// Get menu items excluding the current one
		excludedMenuItems := excludeItem(menuItems, item.ID)

		// Call OpenAI API for recommendations
		suggestions, err := getCrossSellSuggestions(item, excludedMenuItems)
		if err != nil {
			fmt.Printf("Error getting suggestions for item %d: %v\n", item.ID, err)
			continue
		}

		// Store suggestions in database
		storeCrossSellData(item.ID, suggestions, db)

		time.Sleep(2 * time.Second)
	}

	fmt.Println("Cross-sell data population complete!")
}

// Exclude the selected item from the menu list
func excludeItem(menu []MenuItem, itemID uint) []MenuItem {
	var filtered []MenuItem
	for _, item := range menu {
		if item.ID != itemID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Function to call OpenAI API
func getCrossSellSuggestions(selectedItem MenuItem, menu []MenuItem) ([]SuggestedItem, error) {
	// Construct OpenAI prompt
	prompt := fmt.Sprintf(`You are a food recommendation AI. Given the following menu, recommend the top 5 items that best pair with "%s". Rank them from 1 (best fit) to 5.
	Selected Item: %s
	Menu:
	%s
	Respond with only a JSON array of objects in this format:
	[{"id": <id>, "name": "<name>", "priority": <priority>, "description": "<reason why it pairs well>"}]
	`,
		selectedItem.Name,
		selectedItem.Name,
		formatMenu(menu),
	)

	// OpenAI API request
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 500,
	})

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Extract text response
	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}
	content := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	// Clean the response by removing triple backticks and "json" label
	cleanedContent := strings.TrimPrefix(content, "```json")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent) // Remove extra spaces or new lines

	fmt.Printf("Processed AI Response for Item %s: %s\n", selectedItem.Name, cleanedContent) // Debugging

	// Convert AI response JSON into struct
	var suggestions []SuggestedItem
	if err := json.Unmarshal([]byte(cleanedContent), &suggestions); err != nil {
		fmt.Println("Error : ", err.Error())
		return nil, err
	}

	return suggestions, nil
}

// Store cross-sell suggestions in database
func storeCrossSellData(baseItemID uint, suggestions []SuggestedItem, db *gorm.DB) {
	for _, suggestion := range suggestions {
		crossSell := CrossSell{
			BaseItemID:      baseItemID,
			CrossSellItemID: suggestion.ID,
			Priority:        suggestion.Priority,
			Description:     suggestion.Description,
			CreatedAt:       time.Now(),
		}
		if err := db.Create(&crossSell).Error; err != nil {
			fmt.Printf("Failed to insert cross-sell data for item %d: %v\n", baseItemID, err)
		} else {
			fmt.Printf("Inserted cross-sell data: %d -> %d (Priority: %d)\n", baseItemID, suggestion.ID, suggestion.Priority)
		}
	}
}

// Utility: Formats menu for OpenAI input
func formatMenu(menu []MenuItem) string {
	menuStr := ""
	for _, item := range menu {
		menuStr += fmt.Sprintf("- %s (ID: %d)\n", item.Name, item.ID)
	}
	return menuStr
}

func waitForHost(host, port string) error {
	timeOut := time.Second

	if host == "" {
		return errors.Errorf("unable to connect to %v:%v", host, port)
	}

	for i := 0; i < 60; i++ {
		fmt.Printf("waiting for %v:%v ...\n", host, port)
		conn, err := net.DialTimeout("tcp", host+":"+port, timeOut)
		if err == nil {
			fmt.Println("done!")
			conn.Close()
			return nil
		}

		time.Sleep(time.Second)
	}

	return errors.Errorf("timeout attempting to connect to %v:%v", host, port)
}
