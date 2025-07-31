package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
)

type AdClickRequest struct {
	AdvertisementID string `json:"advertisement_id"`
	CafeID          uint   `json:"cafe_id"`
	IsInterested    bool   `json:"is_interested"`
	ClickedCancel   bool   `json:"clicked_cancel"`
}

func (s *Server) RecordUserAdClick(c *fiber.Ctx) error {
	var req AdClickRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate the request
	if req.AdvertisementID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Advertisement ID is required",
		})
	}

	// Create a new ad click record
	adClick := structures.CafeAdvertisementClick{
		CafeID:          req.CafeID,
		AdvertisementID: req.AdvertisementID,
		IsInterested:    req.IsInterested,
		ClickedCancel:   req.ClickedCancel,
		UserID:          uint(c.Locals("userId").(float64)), // Assuming userId is stored in locals
	}

	if err := s.Db.Create(&adClick).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to record ad click",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Ad click recorded successfully",
	})
}
