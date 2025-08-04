package server

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type MenuRequest struct {
	CafeID   string `json:"cafe_id"`
	FoodType string `json:"food_type"`
}

// func (s *Server) GetMenu(c *fiber.Ctx) error {
// 	var req MenuRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	cafeId := req.CafeID
// 	foodType := req.FoodType

// 	if foodType == "1" {
// 		foodType = "veg"
// 	} else if foodType == "2" {
// 		foodType = "non-veg"
// 	}

// 	if cafeId == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "cafe_id is required",
// 		})
// 	}

// 	// Step 1: Fetch categories with counters
// 	var categories []structures.Category
// 	if err := s.Db.Where("cafe_id = ?", cafeId).Find(&categories).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch categories",
// 		})
// 	}

// 	// Step 2: Fetch menu items
// 	var menuItems []structures.MenuItem
// 	fields := []string{
// 		"id", "cafe_id", "category", "name", "description", "short_description",
// 		"price", "is_customizable", "image_url", "video_url", "rating", "total_ratings", "category_id",
// 	}
// 	var err error
// 	if foodType == "" {
// 		err = s.Db.Select(fields).Where("cafe_id = ? AND is_available = true", cafeId).Find(&menuItems).Error
// 	} else {
// 		err = s.Db.Select(fields).Where("cafe_id = ? AND food_type = ? AND is_available = true", cafeId, foodType).Find(&menuItems).Error
// 	}
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch menu items",
// 		})
// 	}

// 	// Step 3: Group menu items by category -> "None"
// 	groupedMenu := make(map[string]map[string][]structures.MenuItem)
// 	for _, item := range menuItems {
// 		if _, ok := groupedMenu[item.Category]; !ok {
// 			groupedMenu[item.Category] = make(map[string][]structures.MenuItem)
// 		}
// 		groupedMenu[item.Category]["None"] = append(groupedMenu[item.Category]["None"], item)
// 	}

// 	// Step 4: Sort categories by ascending counter
// 	sort.Slice(categories, func(i, j int) bool {
// 		return categories[i].Counter < categories[j].Counter
// 	})

// 	// Step 5: Build response in sorted order (as a slice)
// 	orderedResponse := make([]fiber.Map, 0)

// 	for _, cat := range categories {
// 		fmt.Println("Category:", cat.Name, "Counter:", cat.Counter)
// 		if itemsMap, exists := groupedMenu[cat.Name]; exists {
// 			orderedResponse = append(orderedResponse, fiber.Map{
// 				"category": cat.Name,
// 				"items":    itemsMap,
// 			})
// 		}
// 	}

// 	return c.JSON(orderedResponse)
// }

func (s *Server) GetMenu(c *fiber.Ctx) error {
	var req MenuRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}
	if req.CafeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cafe_id is required"})
	}

	// map foodType shortcut
	foodTypeFilter := ""
	switch req.FoodType {
	case "1":
		foodTypeFilter = "veg"
	case "2":
		foodTypeFilter = "non-veg"
	}

	// build SQL
	sql := `
	SELECT
	c.name AS category,
	COALESCE(
		json_agg(
		json_build_object(
			'id'               , m.id,
			'cafe_id'          , m.cafe_id,
			'category_id'      , m.category_id,
			'name'             , m.name,
			'description'      , m.description,
			'short_description', m.short_description,
			'price'            , m.price,
			'is_customizable'  , m.is_customizable,
			'image_url'        , m.image_url,
			'video_url'        , m.video_url,
			'rating'           , m.rating,
			'total_ratings'    , m.total_ratings
		) ORDER BY m.name
		) FILTER (WHERE m.id IS NOT NULL),
		'[]'
	) AS items
	FROM categories c
	LEFT JOIN menu_items m
	ON m.category_id = c.id
	AND m.cafe_id     = c.cafe_id
	AND m.is_available = true
	`

	// optional food_type
	args := []interface{}{}
	if foodTypeFilter != "" {
		sql += " AND m.food_type = ?\n"
		args = append(args, foodTypeFilter)
	}

	sql += `
	WHERE c.cafe_id = ?
	GROUP BY c.name, c.counter
	ORDER BY c.counter ASC;
	`
	args = append(args, req.CafeID)

	// execute
	var out []struct {
		Category string          `json:"category"`
		Items    json.RawMessage `json:"items"`
	}
	if err := s.Db.Raw(sql, args...).Scan(&out).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch menu"})
	}

	return c.JSON(out)
}
