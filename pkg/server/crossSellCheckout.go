package server

import (
	"fmt"
	"net/http"
	"sync"

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
}

func (s *Server) GetCheckoutCrossSells(c *fiber.Ctx) error {
	var req CheckoutCrossSellRequest
	var wg sync.WaitGroup

	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Parallel Fetching
	var bestItems []CheckoutCrossSellResponse
	var fetchErr error

	wg.Add(1)

	// Fetch Bestselling & Best-Rated Desserts and Drinks
	go func() {
		defer wg.Done()
		err := s.Db.Raw(`
			SELECT id AS item_id, name, cm_category AS category, price, image_url, tag, short_description
			FROM menu_items 
			WHERE cm_category IN ('Desserts & Sweets', 'Beverages') 
			AND tag IN ('bestseller', 'bestrated') 
			AND id NOT IN (?) 
			ORDER BY popularity_score DESC
		`, req.ItemIDs).Scan(&bestItems).Error
		if err != nil {
			fetchErr = err
		}
	}()

	wg.Wait()

	// Handle errors
	if fetchErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch checkout cross-sell items",
		})
	}

	fmt.Println("Cross-Sell Items:", bestItems)

	// Return Response
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"result": bestItems,
	})
}
