package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AIRequest struct {
	Query string `json:"query"`
}

// API Handler
func (s *Server) AskMenuAI(c *fiber.Ctx) error {

	var aiRequest AIRequest
	if err := c.BodyParser(&aiRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userQuery := aiRequest.Query
	if userQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}

	// Dynamically include user query in the AI prompt
	prompt := fmt.Sprintf(`You are an AI trained to generate SQL queries based on user prompts. Your goal is to accurately interpret user queries and generate SQL queries based on intent, not just exact keyword matches.

	### **User Query:**
	"%s"

	### **Instructions:**
	- Understand the intent behind the user's query.
	- If it falls into one of the supported scenarios, generate an SQL query.
	- Provide a human-readable explanation of the generated SQL query.

	### **Response Format:**
	Always return a JSON response in the following format:
	json
	{
	"sql": "GENERATED_SQL_QUERY_HERE",
	"response": "A reponse to user prompt eg: Here are the items below 300"
	}
	If the query does not match any supported scenario, return:
	{
	"sql": "",
	"response": "I'm only trained to help you explore menu items based on price, category, tags, or cuisine."
	}
	Supported Scenarios:
	Price-based queries
	Example: "Show me items only below 300"
	SQL: SELECT * FROM menu_items WHERE price < 300
	Category-based queries
	Example: "Show me all the desserts available in this cafe"
	SQL: SELECT * FROM menu_items WHERE category='Desserts'
	Tag-based queries
	Example: "Show me only best selling items in this cafe"
	SQL: SELECT * FROM menu_items WHERE tag='bestseller'
	Cuisine-based queries
	Example: "Show me Italian cuisine available in this cafe"
	SQL: SELECT * FROM menu_items WHERE cuisine='italian'
	Price and category based
	Example: "Show me pizzas under 400 available in this cafe"
	SQL: SELECT * FROM menu_items WHERE category='Pizza' AND price < 400
	Available Categories:
	Biryani, Salads, Grill, Chicken Dishes, Egg Dishes, Main Course, Paneer Dishes, Pizza, Pasta, Lamb & Seafood, Pulao, Mocktails, Cold Coffee, Breakfast, Starters, Conversation Starters, Breads, Tea, Juices, Soups, Curries, Rice, Tacos, Quick Bites, Desserts, Rice & Noodles.

	Available Cuisines:
	Italian, Mexican, Indian, Chinese, Japanese, Mediterranean, Thai, French, American, Korean, Vietnamese, Middle-Eastern, Greek, Spanish.

	Menu Items Schema:
	CREATE TABLE menu_items ( id SERIAL PRIMARY KEY, cafe_id INT NOT NULL, category VARCHAR(50) NOT NULL, sub_category VARCHAR(50), name VARCHAR(100) NOT NULL, description TEXT, price DECIMAL(10,2) NOT NULL, is_customizable BOOLEAN DEFAULT FALSE, food_type VARCHAR(10) NOT NULL, cuisine VARCHAR(50) NOT NULL, dietary_labels VARCHAR(50), spice_level VARCHAR(20), ingredients TEXT, allergens VARCHAR(255), serving_size VARCHAR(50), calories INT, preparation_time INT, discount DECIMAL(5,2), popularity_score FLOAT DEFAULT 0.0, image_url VARCHAR(255), available_from VARCHAR(255), available_till VARCHAR(255), available_all_day BOOLEAN DEFAULT TRUE, is_available BOOLEAN DEFAULT TRUE, tag VARCHAR(255), rating FLOAT DEFAULT 0.0 NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP );
	`, userQuery)

	// Call OpenAI API
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 500,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare AI request"})
	}

	apiKey := s.Config.OPEN_AI_API_KEY

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create AI request"})
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	currentTime := time.Now()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reach AI service"})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read AI response"})
	}

	fmt.Println("Time taken to get response from AI : ", time.Since(currentTime))
	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid AI response format"})
	}

	// Extract AI response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing choices"})
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing message"})
	}

	content, ok := message["content"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI response missing content"})
	}

	// Clean JSON output
	cleanedContent := strings.Trim(content, "`json ")

	// Parse AI response JSON
	var aiResponse map[string]string
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse AI response"})
	}

	// If AI response does not contain SQL, return the response message
	if sqlQuery, exists := aiResponse["sql"]; exists && sqlQuery == "" {
		return c.JSON(fiber.Map{"text": aiResponse["response"], "items": []structures.MenuItem{}})
	}

	fmt.Println("Print SQL query : ", aiResponse["sql"])

	// Execute the generated SQL query
	var menu []structures.MenuItem
	err = s.Db.Raw(aiResponse["sql"]).Scan(&menu).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query execution failed"})
	}

	// Format response text
	responseText := aiResponse["response"]
	if len(menu) == 0 {
		responseText = "This cafe does not have any matching items."
	}

	return c.JSON(fiber.Map{
		"text":  responseText,
		"items": menu,
	})
}
