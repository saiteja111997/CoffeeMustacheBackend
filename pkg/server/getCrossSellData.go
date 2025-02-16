package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type CrossSellRequest struct {
	BaseItemID uint `json:"base_item_id"`
}

func (s *Server) GetCrossSellData(c *fiber.Ctx) error {
	var req CrossSellRequest

	fmt.Println("Printing user id : ", c.Locals("userId").(float64))

	// Parse the JSON request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if BaseItemID is provided
	if req.BaseItemID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "base_item_id is required",
		})
	}

	var crossSells []structures.CrossSell

	// Query the database for cross-sell items based on the input BaseItemID, ordered by Priority DESC
	if err := s.Db.Where("base_item_id = ?", req.BaseItemID).Order("priority DESC").Find(&crossSells).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cross-sell data",
		})
	}
	// Return the list of CrossSellItemIDs as JSON
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"result": crossSells,
	})
}
