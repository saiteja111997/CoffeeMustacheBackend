package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
)

type GetUpsellDataRequest struct {
	CartID string `json:"cart_id"`
	CafeID uint   `json:"cafe_id"`
}

func (s *Server) GetUpsellData(c *fiber.Ctx) error {
	fmt.Println("Fetching Upsell Data ...")

	var getUpsellDataRequest GetUpsellDataRequest

	// Parse request body
	if err := c.BodyParser(&getUpsellDataRequest); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get the total amount of the total cart from the given cart id
	var cart structures.Cart
	if err := s.Db.Model(&structures.Cart{}).Where("cart_id = ?", getUpsellDataRequest.CartID).First(&cart).Error; err != nil {
		fmt.Println("Error fetching total cart amount:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch total cart amount",
		})
	}

	// From the total amount, calculate the difference to reach the next hundered
	nextHundred := ((int(cart.TotalAmount)/100 + 1) * 100) - int(cart.TotalAmount)
	fmt.Println("Next hundred amount to reach:", nextHundred)

	// Add 100 to the next hundred amount to get the upsell amount
	upsellAmount := nextHundred + 100
	fmt.Println("Upsell amount:", upsellAmount)

	// For every 50 rupees we need to give 10 mustaches
	mustachesToGive := (upsellAmount / 50) * 10
	fmt.Println("Mustaches to give:", mustachesToGive)

	// Generate upsell id from ksuid
	upsellID := ksuid.New().String()

	// Insert into the upsell table
	upsellData := structures.UpsellData{
		CartID:          getUpsellDataRequest.CartID,
		CafeID:          getUpsellDataRequest.CafeID,
		CurrentAmount:   cart.TotalAmount,
		TargetAmount:    float64(nextHundred),
		MustachesToGive: uint(mustachesToGive),
		UpsellID:        upsellID,
	}

	if err := s.Db.Create(&upsellData).Error; err != nil {
		fmt.Println("Error inserting upsell data:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to insert upsell data",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"upsell_amount":   upsellAmount,
		"mustaches_given": mustachesToGive,
		"upsell_id":       upsellID,
		"message":         "Upsell data fetched successfully",
		"status":          "success",
	})
}
