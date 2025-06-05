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

type menuItemInput struct {
	ID               uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Category         string  `gorm:"type:varchar(50);not null" json:"category"`
	SubCategory      string  `gorm:"type:varchar(50)" json:"sub_category"`
	Name             string  `gorm:"type:varchar(100);not null" json:"name"`
	ShortDescription string  `gorm:"type:varchar(255)" json:"short_description"`
	Price            float64 `gorm:"type:decimal(10,2);not null" json:"price"`
}

func (s *Server) RunCuratedCartsJob(c *fiber.Ctx) error {
	// Fetch all cafe IDs
	var cafeIDs []uint
	if err := s.Db.Table("cafes").Pluck("id", &cafeIDs).Error; err != nil {
		log.Println("❌ Failed to fetch cafe IDs:", err)
		return err
	}

	for _, cafeID := range cafeIDs {
		var menuItems []menuItemInput
		if err := s.Db.Table("menu_items").Where("cafe_id = ?", cafeID).Find(&menuItems).Error; err != nil {
			log.Printf("❌ Failed to fetch menu items for cafe %d: %v\n", cafeID, err)
			continue
		}
		if len(menuItems) == 0 {
			log.Printf("No menu items found for cafe %d, skipping...\n", cafeID)
			continue
		}

		prompt := fmt.Sprintf(`Generate 9 curated carts (3 morning, 3 noon, 3 night) for a cafe. 
		Here are the menu items: %v. 
		Each cart should:
		- Have a catchy name
		- Contain a well-balanced mix of items across categories (e.g., a drink, a main, and a dessert)
		- Ensure a reasonable total price
		- Be output as JSON: [{"name": "Morning Bliss", "item_ids": [1,2,3], "time_of_day": "morning"}]
		- In the response just give the list of items, nothing else`, menuItems)

		requestBody, _ := json.Marshal(map[string]interface{}{
			"model":      "gpt-4o",
			"messages":   []map[string]string{{"role": "user", "content": prompt}},
			"max_tokens": 2000,
		})

		// fmt.Println("Request Body for OpenAI:", string(requestBody))

		req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
		req.Header.Set("Authorization", "Bearer "+s.Config.OPEN_AI_API_KEY)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ OpenAI API call failed for cafe %d: %v\n", cafeID, err)
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		// Parse OpenAI response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("❌ Failed to unmarshal OpenAI response for cafe %d: %v\n", cafeID, err)
			continue
		}

		fmt.Println("OpenAI Response for Cafe ID:", result)

		// Extract text response
		choices, ok := result["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			log.Printf("No response from OpenAI for cafe %d\n", cafeID)
			continue
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
			log.Printf("Error unmarshalling curated carts for cafe %d: %v\n", cafeID, err)
			continue
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
				if err := s.Db.Where("id =? AND cafe_id = ?", itemID, cafeID).First(&item).Error; err != nil {
					log.Println("Failed to fetch item:", err)
					continue
				}
				totalAmount += item.Price
			}

			// Calculate discount amount
			discountAmount, discountPercent := 0.0, 0.0
			if totalAmount > 1200 {
				discountAmount = 0.10 * totalAmount
				discountPercent = 10
			} else {
				discountAmount = 0.15 * totalAmount
				discountPercent = 15
			}

			curatedCart := structures.CuratedCart{
				CafeID:          cafeID,
				Name:            cart.Name,
				TimeOfDay:       cart.TimeOfDay,
				Date:            time.Now(),
				Source:          "ai",
				CartTotal:       totalAmount,
				DiscountedTotal: totalAmount - discountAmount,
				DiscountPercent: discountPercent,
				ButtonActions:   0,
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
		log.Printf("✅ Curated carts successfully generated and stored for cafe %d.\n", cafeID)
	}
	return nil
}
