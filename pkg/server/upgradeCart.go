package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UpgradeCartRequest struct {
	CartID  string `json:"cart_id"`
	ItemIDs []uint `json:"item_ids"` // List of item IDs currently in the cart
}

type UpgradeCartResponse struct {
	ItemID          uint    `json:"item_id"`
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	Price           float64 `json:"price"`
	UserReason      string  `json:"user_reason"`
	ReferenceReason string  `json:"reference_reason"`
	ImageURL        string  `json:"image_url"`
}

func (s *Server) UpgradeCart(c *fiber.Ctx) error {
	var req UpgradeCartRequest
	var wg sync.WaitGroup

	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate input
	if req.CartID == "" || len(req.ItemIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cart_id and item_ids are required",
		})
	}

	// Fetch Menu (excluding cart items) & Cart Items
	var cartItems []UpgradeCartResponse
	var menuItems []UpgradeCartResponse
	var fetchCartErr, fetchMenuErr error

	wg.Add(2)

	// Get Cart Items (for AI input)
	go func() {
		defer wg.Done()
		err := s.Db.Raw(`
			SELECT id AS item_id, name, category, price
			FROM menu_items 
			WHERE id IN (?)
		`, req.ItemIDs).Scan(&cartItems).Error
		if err != nil {
			fetchCartErr = err
		}
	}()

	// Get Full Menu (excluding cart items)
	go func() {
		defer wg.Done()
		err := s.Db.Raw(`
			SELECT id AS item_id, name, category, price
			FROM menu_items 
			WHERE id NOT IN (?)
		`, req.ItemIDs).Scan(&menuItems).Error
		if err != nil {
			fetchMenuErr = err
		}
	}()

	wg.Wait()

	// Handle errors
	if fetchCartErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cart items",
		})
	}
	if fetchMenuErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu items",
		})
	}

	// Convert cart items & menu items to JSON for AI
	cartJSON, _ := json.Marshal(cartItems)
	menuJSON, _ := json.Marshal(menuItems)

	// Prepare AI Prompt
	prompt := fmt.Sprintf(`
	You are an AI recommendation system for a cafe. Your task is to suggest **one** perfect complementary item 
	that enhances the user's cart experience based on their current order.

	### User's Cart:
	%s

	### Available Menu (excluding cart items):
	%s

	**Recommendation Rules:**
	1. The suggested item should balance the cart, filling missing categories if needed.
	2. Consider price: If the cart is high-cost, suggest a complementary premium item; if low-cost, suggest an affordable addition.
	3. Pick an item that users typically order with this combination.
	4. Give a persuasive short one liner, as user_reason for the recommendation based on the user's cart.
	5. Provide a reference_reason for the recommendation based on the menu data, for us to later analyze the AI's decision-making.
	6. Output **only** JSON with the following fields: `+"`item_id, name, category, price, user_reason, reference_reason`"+`.


	Now, recommend the best complementary item in JSON format:
	`, string(cartJSON), string(menuJSON))

	// fmt.Println("AI Prompt:", prompt)

	// Call OpenAI API
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 1000,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to prepare AI request",
		})
	}

	apiKey := s.Config.OPEN_AI_API_KEY

	reqAI, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create AI request",
		})
	}

	reqAI.Header.Set("Authorization", "Bearer "+apiKey)
	reqAI.Header.Set("Content-Type", "application/json")

	currentTime := time.Now()
	client := &http.Client{}
	resp, err := client.Do(reqAI)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reach AI service",
		})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read AI response",
		})
	}

	fmt.Println("Time taken to get response from AI:", time.Since(currentTime))

	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid AI response format",
		})
	}

	// Extract AI response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI response missing choices",
		})
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI response missing message",
		})
	}

	content, ok := message["content"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI response missing content",
		})
	}

	fmt.Println("AI Response:", content)

	// Clean JSON output
	cleanedContent := strings.Trim(content, "`json ")

	// Parse the cleaned JSON into struct
	var recommendedItem UpgradeCartResponse
	if err := json.Unmarshal([]byte(cleanedContent), &recommendedItem); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse AI response",
		})
	}

	wg.Add(2)

	// Insert AI suggestion into update_cart_result table
	var insertErr, fetchImageErr error
	go func() {
		defer wg.Done()
		newEntry := structures.UpdateCartResult{
			CartID:                req.CartID,
			SuggestedItemID:       recommendedItem.ItemID,
			SuggestedItemName:     recommendedItem.Name,
			SuggestedItemCategory: recommendedItem.Category,
			SuggestedItemPrice:    recommendedItem.Price,
			AIResponse:            cleanedContent, // Storing full AI response as JSONB
			UserReason:            recommendedItem.UserReason,
			ReferenceReason:       recommendedItem.ReferenceReason,
			UserAction:            "pending",
		}

		if err := s.Db.Create(&newEntry).Error; err != nil {
			insertErr = err
		}
	}()

	// Get image url from menu items and populate in the response
	go func() {
		defer wg.Done()
		var menuItem structures.MenuItem
		if err := s.Db.Where("id =?", recommendedItem.ItemID).First(&menuItem).Error; err != nil {
			fetchImageErr = err
			return
		}
		recommendedItem.ImageURL = menuItem.ImageURL // Populate image URL from menu items table
	}()

	wg.Wait()

	if insertErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to log AI suggestion",
		})
	}

	// Handle errors
	if fetchImageErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu item",
		})
	}

	// Return AI-suggested item
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": recommendedItem,
	})
}
