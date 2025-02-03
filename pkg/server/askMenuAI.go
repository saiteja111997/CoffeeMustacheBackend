package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// API Handler
func (s *Server) AskMenuAI(c *fiber.Ctx) error {
	userQuery := c.FormValue("query")
	if userQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	// Fetch menu items
	var menu []structures.MenuItem
	err := s.Db.Find(&menu).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Construct OpenAI prompt
	prompt := fmt.Sprintf(`You are a food recommendation AI trained to answer only about the menu. Given the menu items below, answer the user’s question accordingly.

    **Rules:**
    - If the question is unrelated to food, return: 
      {"response": "I'm only trained to help you explore the menu. Please ask me something related to food or drinks.", "items": []}
    
    - If the user asks about available items, return:
      {"response": "Here is what we have:", "items": [{"id": 1, "name": "Item Name", "price": 450, "rating": 4.5}]}

    - If the user asks for best sellers, return:
      {"response": "Here are our best sellers:", "items": [{"id": 5, "name": "Best Seller", "price": 400, "rating": 4.8}]}

    - If the user asks about budget, return items under the given budget:
      {"response": "Here are items under ₹X:", "items": [{"id": 6, "name": "Budget Meal", "price": 350, "rating": 4.2}]}

    - If the user asks for pairing or cross-sell, return:
      {"response": "Here are some great pairings:", "items": [{"id": 7, "name": "Cappuccino", "price": 200, "rating": 4.6}]}

    - If the user asks for combos, return:
      {"response": "Here are some great meal combos:", "combos": [{"name": "Meal for Two", "items": [{"id": 3, "name": "Pizza", "price": 350, "rating": 4.6}, {"id": 4, "name": "Iced Latte", "price": 150, "rating": 4.3}], "total_price": 500}]}

    **User Query:** %s

    **Menu Items:**
    %s
    `, userQuery, formatMenuInput(menu))

	// Call OpenAI API
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 5000,
	})

	apiKey := s.Config.OPEN_AI_API_KEY

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid response from AI"})
	}

	// Extract AI response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response error"})
	}

	content := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	fmt.Println("Content: ", content)

	// Clean JSON output
	cleanedContent := strings.TrimPrefix(content, "```json")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	// Convert AI response JSON into struct
	var aiSuggestions map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedContent), &aiSuggestions); err != nil {
		fmt.Println("Error parsing AI response:", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse AI response"})
	}

	// Return AI's structured response
	return c.JSON(aiSuggestions)
}

// Utility: Formats menu for OpenAI input
func formatMenuInput(menu []structures.MenuItem) string {
	menuStr := ""
	for _, item := range menu {
		menuStr += fmt.Sprintf("- %s (ID: %d, Price: ₹%.2f, Rating: %.1f, Tag: %s)\n",
			item.Name, item.ID, item.Price, item.Rating, item.Tag)
	}
	return menuStr
}
