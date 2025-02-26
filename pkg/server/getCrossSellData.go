package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CrossSellRequest struct {
	BaseItemID uint   `json:"base_item_id"`
	CartID     string `json:"cart_id"`
}

type CrossSellResponse struct {
	CrossSellID       uint    `json:"cross_sell_id"`
	BaseItemID        uint    `json:"base_item_id"`
	CrossSellItemID   uint    `json:"cross_sell_item_id"`
	CrossSellCategory string  `json:"cross_sell_category"`
	Priority          int     `json:"priority"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	Category          string  `json:"category"`
	ImageURL          string  `json:"image_url"`
	ItemName          string  `json:"item_name"`
}

func (s *Server) GetCrossSellData(c *fiber.Ctx) error {
	var req CrossSellRequest
	var wg sync.WaitGroup

	// Track execution time
	startTime := time.Now()

	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if this is a keep-warm event
	if req.BaseItemID == 0 && req.CartID == "keep-warm" {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "Lambda kept warm",
		})
	}

	// Validate required fields
	if req.BaseItemID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "base_item_id is required",
		})
	}

	// If no cart_id is provided, fetch enriched cross-sell items directly
	if req.CartID == "" {
		var crossSellItems []CrossSellResponse
		if err := s.Db.Raw(`
			SELECT 
				cs.id AS cross_sell_id,
				cs.base_item_id,
				cs.cross_sell_item_id,
				cs.cross_sell_category,
				cs.priority,
				cs.description,
				mi.price,
				mi.category,
				mi.image_url,
				mi.name AS item_name
			FROM cross_sells cs
			JOIN menu_items mi ON cs.cross_sell_item_id = mi.id
			WHERE cs.base_item_id = ?
			ORDER BY cs.priority DESC
		`, req.BaseItemID).Scan(&crossSellItems).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch cross-sell data",
			})
		}

		fmt.Println("Execution Time:", time.Since(startTime)) // Log time
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"result": crossSellItems,
		})
	}

	// Use parallel execution for fetching cart items and cross-sell items
	var cartItemIDs []uint
	var crossSellItems []CrossSellResponse
	var fetchCartErr, fetchCrossSellErr error

	// Start parallel execution
	wg.Add(2)

	// Goroutine 1: Fetch cart items (Optimized using raw SQL)
	go func() {
		defer wg.Done()
		err := s.Db.Raw(`
			SELECT item_id FROM cart_items 
			WHERE cart_id = ?
		`, req.CartID).Pluck("item_id", &cartItemIDs).Error
		if err != nil {
			fetchCartErr = err
		}
	}()

	// Goroutine 2: Fetch cross-sell items with menu item details
	go func() {
		defer wg.Done()
		err := s.Db.Raw(`
			SELECT 
				cs.id AS cross_sell_id,
				cs.base_item_id,
				cs.cross_sell_item_id,
				cs.cross_sell_category,
				cs.priority,
				cs.description,
				mi.price,
				mi.category,
				mi.image_url,
				mi.name AS item_name
			FROM cross_sells cs
			JOIN menu_items mi ON cs.cross_sell_item_id = mi.id
			WHERE cs.base_item_id = ?
			ORDER BY cs.priority DESC
		`, req.BaseItemID).Scan(&crossSellItems).Error
		if err != nil {
			fetchCrossSellErr = err
		}
	}()

	// Wait for both queries to complete
	wg.Wait()

	// Check for errors
	if fetchCartErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cart items",
		})
	}
	if fetchCrossSellErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cross-sell data",
		})
	}

	// Filter cross-sell items in memory
	cartItemSet := make(map[uint]bool)
	for _, id := range cartItemIDs {
		cartItemSet[id] = true
	}

	var filteredCrossSellItems []CrossSellResponse
	for _, crossSell := range crossSellItems {
		if !cartItemSet[crossSell.CrossSellItemID] {
			filteredCrossSellItems = append(filteredCrossSellItems, crossSell)
		}
	}

	// Log execution time
	fmt.Println("Execution Time:", time.Since(startTime))

	// Return the filtered cross-sell items
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"result": filteredCrossSellItems,
	})
}
