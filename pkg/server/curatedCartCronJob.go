package server

import (
	"bytes"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

func (s *Server) RunCuratedCartsJob(c *fiber.Ctx) error {

	var menuItems []structures.MenuItem
	if err := s.Db.Table("menu_items").Find(&menuItems).Error; err != nil {
		log.Println("❌ Failed to fetch menu items:", err)
		return err
	}

	prompt := fmt.Sprintf(`Generate 9 curated carts (3 morning, 3 noon, 3 night) for a cafe. 
	Here are the menu items: %v. 
	Each cart should:
	- Have a catchy name
	- Contain a well-balanced mix of items across categories (e.g., a drink, a main, and a dessert)
	- Ensure a reasonable total price
	- Be output as JSON: [{"name": "Morning Bliss", "item_ids": [1,2,3], "time_of_day": "morning"}]`, menuItems)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o",
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 2000,
	})

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer "+s.Config.OPEN_AI_API_KEY)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ OpenAI API call failed:", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Parse OpenAI response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	fmt.Println("Printing response from OpenAI : ", result)

	// Extract text response
	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		return fmt.Errorf("no response from OpenAI")
	}
	content := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	// Clean the response
	cleanedContent := strings.TrimPrefix(content, "```json")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	var curatedCarts []struct {
		Name      string               `json:"name"`
		ItemIDs   []uint               `json:"item_ids"`
		TimeOfDay structures.TimeOfDay `json:"time_of_day"`
	}
	if err := json.Unmarshal([]byte(cleanedContent), &curatedCarts); err != nil {
		fmt.Println("Error:", err.Error())
		return err
	}

	for _, cart := range curatedCarts {

		// Convert `ItemIDs` to JSON before saving
		itemIDsJSON, err := json.Marshal(cart.ItemIDs)
		if err != nil {
			log.Println("❌ Failed to convert ItemIDs to JSON:", err)
			continue
		}

		// Calculate cart total amount
		var totalAmount float64
		for _, itemID := range cart.ItemIDs {
			var item structures.MenuItem
			if err := s.Db.Where("id =?", itemID).First(&item).Error; err != nil {
				log.Println("Failed to fetch item:", err)
				continue
			}
			totalAmount += item.Price
		}

		// Calculate dicount amount. If the cart value is more than 1200 give 10% of the total amount else 15% of the total amount
		discountAmount, discountPercent := 0.0, 0.0
		if totalAmount > 1200 {
			discountAmount = 0.10 * totalAmount
			discountPercent = 10
		} else {
			discountAmount = 0.15 * totalAmount
			discountPercent = 15
		}

		curatedCart := structures.CuratedCart{
			CafeID:          1, // Replace with actual CafeID if needed
			Name:            cart.Name,
			TimeOfDay:       cart.TimeOfDay,
			Date:            time.Now(),
			Source:          "ai",
			CartTotal:       totalAmount,
			DiscountedTotal: totalAmount - discountAmount,
			DiscountPercent: discountPercent,
			ButtonActions:   0, // Replace with actual button actions if needed, e.g., 3 for 3-button action carts
			ItemIDs:         datatypes.JSON(itemIDsJSON),
		}
		if err := s.Db.Table("curated_carts").Create(&curatedCart).Error; err != nil {
			log.Println("❌ Failed to insert curated cart:", err)
			continue
		}

		for i, itemID := range cart.ItemIDs {
			curatedCartItem := structures.CuratedCartItem{
				CartID:   curatedCart.ID,
				ItemID:   itemID,
				Priority: i,
			}
			if err := s.Db.Table("curated_cart_items").Create(&curatedCartItem).Error; err != nil {
				log.Println("❌ Failed to insert curated cart item:", err)
			}
		}
	}

	log.Println("✅ Curated carts successfully generated and stored.")
	return nil
}
