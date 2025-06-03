package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MenuRequest struct {
	CafeID   string `json:"cafe_id"`
	FoodType string `json:"food_type"`
}

func (s *Server) GetMenu(c *fiber.Ctx) error {

	var req MenuRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	cafeId := req.CafeID
	foodType := req.FoodType

	fmt.Println("Cafe Id : ", cafeId)
	fmt.Println("Printing food type : ", foodType)

	if foodType == "1" {
		foodType = "veg"
	} else if foodType == "2" {
		foodType = "non-veg"
	}

	// Validate input
	if cafeId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cafe_id is required",
		})
	}

	var menuItems []structures.MenuItem

	startTime := time.Now()

	// Fetch only the required fields from the database
	var err error
	fields := []string{
		"id", "cafe_id", "category", "name", "description", "short_description",
		"price", "is_customizable", "image_url", "video_url", "rating", "total_ratings",
	}
	if foodType == "" {
		err = s.Db.Select(fields).Where("cafe_id = ? AND is_available = true", cafeId).Find(&menuItems).Error
	} else {
		err = s.Db.Select(fields).Where("cafe_id = ? AND food_type = ? AND is_available = true", cafeId, foodType).Find(&menuItems).Error
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch menu items",
		})
	}

	currentTime := time.Now()

	fmt.Println("Printing time difference : ", currentTime.Sub(startTime))

	fmt.Println("No of Menu items : ", len(menuItems))

	// Group menu items by category and sub-category
	groupedMenu := make(map[string]map[string][]structures.MenuItem)
	for _, item := range menuItems {
		if _, ok := groupedMenu[item.Category]; !ok {
			groupedMenu[item.Category] = make(map[string][]structures.MenuItem)
		}

		groupedMenu[item.Category]["None"] = append(groupedMenu[item.Category]["None"], item)

	}

	return c.JSON(groupedMenu)

}
