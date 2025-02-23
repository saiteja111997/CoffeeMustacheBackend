package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
)

// GetItemAudio handles the request to get the audio URL for a specific menu item.
// It expects a JSON body with cafe_id, item_id, and language.
func (s *Server) GetItemAudio(c *fiber.Ctx) error {
	// Parse request body
	var request struct {
		CafeID uint `json:"cafe_id"`
		ItemID uint `json:"item_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Query the database for the menu item with matching CafeID and ItemID.
	var menuItem structures.MenuItem
	if err := s.Db.Where("cafe_id = ? AND id = ?", request.CafeID, request.ItemID).First(&menuItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Menu item not found",
		})
	}

	// Return the audio URL.
	return c.JSON(fiber.Map{
		"audio_url": menuItem.AudioURL,
	})
}
