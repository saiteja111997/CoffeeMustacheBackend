package server

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type CheckoutCrossSellRequest struct {
	ItemIDs []uint `json:"item_ids"` // Array of item IDs already in the cart
}

type CheckoutCrossSellResponse struct {
	ShortDescription string  `json:"short_description"`
	ItemID           uint    `json:"item_id"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	Price            float64 `json:"price"`
	ImageURL         string  `json:"image_url"`
	Tag              string  `json:"tag"`
	DiscountedPrice  float64 `json:"discounted_price"`
	DiscountPercent  float64 `json:"discount_percent"`
}

func (s *Server) GetCheckoutCrossSells(c *fiber.Ctx) error {
	var req CheckoutCrossSellRequest

	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Parallel Fetching
	var bestItems []CheckoutCrossSellResponse
	var fetchErr error
	err := s.Db.Raw(`
		SELECT id AS item_id, name, cm_category AS category, price, image_url, tag, short_description, 
			discount as discount_percent, 
			ROUND(price - (price * discount / 100)) AS discounted_price
		FROM menu_items 
		WHERE cm_category IN ('Desserts & Sweets', 'Beverages') 
		AND (tag @> '["bestseller"]' OR tag @> '["bestrated"]')
		AND id NOT IN (?)
		ORDER BY popularity_score DESC
`, req.ItemIDs).Scan(&bestItems).Error

	if err != nil {
		fetchErr = err
	}

	// Handle errors
	if fetchErr != nil {
		fmt.Println("Error fetching checkout cross-sell items: ", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch checkout cross-sell items",
		})
	}

	// Return Response
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"result": bestItems,
	})
}
