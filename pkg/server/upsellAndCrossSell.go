package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
)

// API Handler
func (s *Server) GetUpsellAndCrossSell(c *fiber.Ctx) error {

	var input struct {
		ItemID string `json:"item_id"`
	}

	// Parse JSON input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate input
	if input.ItemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "item_id is required",
		})
	}

	itemId := input.ItemID

	// Fetch Upsells (Customizations)
	var upsells []structures.ItemCustomization
	err := s.Db.Where("menu_item_id = ?", itemId).Find(&upsells).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch upsells",
		})
	}

	upsellMap := make(map[string][]structures.UpsellCategory)
	for _, upsell := range upsells {
		upsellMap[upsell.CustomizationType] = append(upsellMap[upsell.CustomizationType], structures.UpsellCategory{
			Name:           upsell.OptionName,
			AdditionalCost: upsell.AdditionalCost,
		})
	}

	// Fetch Cross-Sell Items
	var crossSells []structures.CrossSell
	err = s.Db.Where("base_item_id = ?", itemId).Find(&crossSells).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cross-sells",
		})
	}

	crossSellMap := make(map[string][]structures.CrossSellCategory)
	for _, crossSell := range crossSells {
		var item structures.MenuItem
		err = s.Db.Where("id = ?", crossSell.CrossSellItemID).First(&item).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch cross-sell item details",
			})
		}

		crossSellMap[crossSell.CrossSellCategory] = append(crossSellMap[crossSell.CrossSellCategory], structures.CrossSellCategory{
			Name:     item.Name,
			Priority: crossSell.Priority,
		})
	}

	// Build Response
	response := structures.Response{}
	response.Result.Upsell = upsellMap
	response.Result.CrossSell = crossSellMap

	return c.JSON(response)
}
