package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetCuratedCart(c *fiber.Ctx) error {
	// Parse cafe ID from request body
	type CuratedCartRequest struct {
		CafeID uint `json:"cafe_id"`
	}

	var req CuratedCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Get the current date and time in IST
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load location",
		})
	}
	currentTime := time.Now().In(location)
	currentDate := currentTime.Format("2006-01-02")
	currentHour := currentTime.Hour()

	// Determine time of day
	var timeOfDay structures.TimeOfDay
	switch {
	case currentHour < 12:
		timeOfDay = structures.Morning
	case currentHour >= 12 && currentHour < 17:
		timeOfDay = structures.Afternoon
	default:
		timeOfDay = structures.Night
	}

	// Fetch curated carts for the given cafe, current date, and time of day
	var curatedCarts []structures.CuratedCart
	if err := s.Db.Where("cafe_id = ? AND date = ? AND time_of_day = ?", req.CafeID, currentDate, timeOfDay).Find(&curatedCarts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch curated carts",
		})
	}

	type CuratedCartItemDetailResponse struct {
		ItemID         uint    `json:"item_id"`
		Name           string  `json:"name"`
		Price          float64 `json:"price"`
		ImageURL       string  `json:"image_url"`
		IsCustomizable bool    `json:"is_customizable"`
	}

	// Prepare a response with carts and their items, including item name and price
	type CuratedCartResponse struct {
		ID              uint                            `json:"id"`
		CafeID          uint                            `json:"cafe_id"`
		Name            string                          `json:"name"`
		TimeOfDay       string                          `json:"time_of_day"`
		Date            string                          `json:"date"`
		Source          string                          `json:"source"`
		CartTotal       float64                         `json:"cart_total"`
		DiscountedTotal float64                         `json:"discounted_total"`
		DiscountPercent float64                         `json:"discount_percent"`
		Items           []CuratedCartItemDetailResponse `json:"items"`
	}

	var response []CuratedCartResponse
	for _, cart := range curatedCarts {
		// Fetch items for each curated cart with item name and price
		var curatedItems []structures.CuratedCartItem
		if err := s.Db.Where("cart_id = ?", cart.ID).Find(&curatedItems).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch items for curated cart",
			})
		}

		var itemDetails []CuratedCartItemDetailResponse
		for _, item := range curatedItems {
			// Fetch item name and price from the MenuItem table
			var menuItem structures.MenuItem
			if err := s.Db.Select("id, name, price, image_url, is_customizable").Where("id = ?", item.ItemID).First(&menuItem).Error; err == nil {
				itemDetails = append(itemDetails, CuratedCartItemDetailResponse{
					ItemID:         menuItem.ID,
					Name:           menuItem.Name,
					Price:          menuItem.Price,
					ImageURL:       menuItem.ImageURL,
					IsCustomizable: menuItem.IsCustomizable,
				})
			}
		}

		// Add cart and items to response
		response = append(response, CuratedCartResponse{
			ID:              cart.ID,
			CafeID:          cart.CafeID,
			Name:            cart.Name,
			TimeOfDay:       string(cart.TimeOfDay),
			Date:            cart.Date.Format("2006-01-02"),
			Source:          cart.Source,
			CartTotal:       cart.CartTotal,
			DiscountedTotal: cart.DiscountedTotal,
			DiscountPercent: cart.DiscountPercent,
			Items:           itemDetails,
		})
	}

	// Return the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Curated carts fetched successfully",
		"curated_carts": response,
	})
}
