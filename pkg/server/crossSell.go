package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
)

// API Handler for Cross-Selling
func (s *Server) CrossSellItem(c *fiber.Ctx) error {

	itemId := c.FormValue("item_id")

	// Validate input
	if itemId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "itemId is required",
		})
	}

	itemIdInt, err := strconv.Atoi(itemId)

	// Get the selected item details
	var selectedItem structures.MenuItem
	if err := s.Db.First(&selectedItem, itemIdInt).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	// Fetch menu excluding the selected item
	menu, err := getMenuItemsExcluding(itemIdInt, s.Db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Call OpenAI API to get recommendations
	suggestions, err := getCrossSellSuggestions(selectedItem, menu, s.Config.OPEN_AI_API_KEY)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return response
	return c.JSON(fiber.Map{
		"suggestions": structures.CrossSellResponse{Suggestions: suggestions},
	})
}

// Fetch all menu items excluding the selected one
func getMenuItemsExcluding(itemID int, db *gorm.DB) ([]structures.MenuItem, error) {
	var menu []structures.MenuItem
	if err := db.Where("id != ?", itemID).Find(&menu).Error; err != nil {
		return nil, err
	}
	return menu, nil
}

// Function to call OpenAI API
func getCrossSellSuggestions(selectedItem structures.MenuItem, menu []structures.MenuItem, apiKey string) ([]structures.SuggestedItem, error) {
	// Construct OpenAI prompt
	prompt := fmt.Sprintf(`You are a food recommendation AI. Given the following menu, recommend the top 5 items that best pair with "%s". Rank them from 1 (best fit) to 5.
	Selected Item: %s
	Menu:
	%s
	Respond with only a JSON array of objects in this format:
	[{"id": <id>, "name": "<name>", "priority": <priority>}]
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

	// fmt.Println("OpenAI response : ", result)

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

	fmt.Println("Cleaned JSON Content:", cleanedContent) // Debugging line

	// Convert AI response JSON into struct
	var suggestions []structures.SuggestedItem
	if err := json.Unmarshal([]byte(cleanedContent), &suggestions); err != nil {
		fmt.Println("Error : ", err.Error())
		return nil, err
	}

	return suggestions, nil
}

// Utility: Formats menu for OpenAI input
func formatMenu(menu []structures.MenuItem) string {
	menuStr := ""
	for _, item := range menu {
		menuStr += fmt.Sprintf("- %s (ID: %d)\n", item.Name, item.ID)
	}
	return menuStr
}
