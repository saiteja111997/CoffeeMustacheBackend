package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm" // For jsonb handling in Go
)

type TopPicksRequest struct {
	Tag string `json:"tag"`
}

func (s *Server) GetFilteredList(c *fiber.Ctx) error {
	var req TopPicksRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Request",
		})
	}

	// Tags : “price 200-400”, “bestrated”, “trending”, “bestsellers”, "toppicks" check if the tag is valid
	validTags := map[string]bool{
		"price 200-400": true,
		"bestrated":     true,
		"trending":      true,
		"bestseller":    true,
		"toppicks":      true,
	}

	if !validTags[req.Tag] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tag",
		})
	}

	var items []structures.MenuItem

	// if the tag is price 200-400, run a sql query on menu items table to get all items with price between 200 and 400 and return the response directly.
	if req.Tag == "price 200-400" {
		if err := s.Db.Where("price BETWEEN 200 AND 400").Find(&items).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "No items found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}
		return c.JSON(fiber.Map{
			"items": items,
		})
	}

	// Fetch all items and their tags
	if err := s.Db.Select("*").Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No items found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var filteredItems []structures.MenuItem
	// Loop over the fetched items and check if the tag exists in the 'tags' field
	for _, item := range items {
		var tags []string
		// Parse the jsonb tags column into a Go slice of strings
		if err := json.Unmarshal(item.Tag, &tags); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error processing tags data",
			})
		}

		// Check if the requested tag exists in the tags slice
		for _, tag := range tags {
			if tag == req.Tag {
				// If the tag matches, add it to the filtered list
				filteredItems = append(filteredItems, item)
				break
			}
		}
	}

	if len(filteredItems) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No items found for the given tag",
		})
	}

	// Return the filtered items as JSON response
	return c.JSON(fiber.Map{
		"items": filteredItems,
	})
}
