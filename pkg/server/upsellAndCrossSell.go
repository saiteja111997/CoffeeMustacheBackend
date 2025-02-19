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

	// Organize upsells into a map
	upsellMap := make(map[string][]structures.UpsellCategory)
	for _, upsell := range upsells {
		upsellMap[upsell.CustomizationType] = append(upsellMap[upsell.CustomizationType], structures.UpsellCategory{
			Name:            upsell.OptionName,
			AdditionalCost:  upsell.AdditionalCost,
			CustomizationID: upsell.ID,
		})
	}

	// Fetch Cross-Sell Items with Menu Details (Using JOIN for efficiency)
	var crossSells []structures.CrossSellCategory
	err = s.Db.Raw(`
		SELECT 
			cs.cross_sell_category,
			mi.id AS item_id,
			mi.name AS name,
			mi.price AS price,
			cs.priority
		FROM cross_sells cs
		JOIN menu_items mi ON cs.cross_sell_item_id = mi.id
		WHERE cs.base_item_id = ?
		ORDER BY cs.priority DESC
	`, itemId).Scan(&crossSells).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cross-sell items",
		})
	}

	// Organize cross-sell items into a map
	crossSellMap := make(map[string][]structures.CrossSellCategory)
	for _, crossSell := range crossSells {
		crossSellMap[crossSell.CrossSellCategory] = append(crossSellMap[crossSell.CrossSellCategory], crossSell)
	}

	// Build Response
	response := structures.Response{}
	response.Result.Upsell = upsellMap
	response.Result.CrossSell = crossSellMap

	return c.JSON(response)
}
