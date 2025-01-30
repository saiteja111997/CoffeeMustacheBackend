package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) UpsellItem(c *fiber.Ctx) error {
	itemId := c.FormValue("item_id")

	// Validate input
	if itemId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "itemId is required",
		})
	}

	// Parse item ID
	var menuItem structures.MenuItem
	if err := s.Db.First(&menuItem, itemId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Item not found",
		})
	}

	// Fetch customizations related to the item
	var customizations []structures.ItemCustomization
	if err := s.Db.Where("menu_item_id = ?", itemId).Find(&customizations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch customizations",
		})
	}

	// Group customizations by category
	customizationMap := make(map[string][]structures.CustomizationItem)
	for _, customization := range customizations {
		item := structures.CustomizationItem{
			ItemName:       customization.OptionName,
			AdditionalCost: customization.AdditionalCost,
		}
		customizationMap[customization.CustomizationType] = append(customizationMap[customization.CustomizationType], item)
	}

	// Convert map to structured response
	var responseCategories []structures.CustomizationCategory
	for category, items := range customizationMap {
		responseCategories = append(responseCategories, structures.CustomizationCategory{
			Category: category,
			Items:    items,
		})
	}

	// Construct response
	response := structures.CustomizationResponse{
		ItemID:         menuItem.ID,
		Customizations: responseCategories,
	}

	return c.JSON(response)
}
