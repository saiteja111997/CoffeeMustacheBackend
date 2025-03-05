package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
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

	// Tags : “price 200-400”, “bestrated”, “trending”, “bestsellers” check if the tag is valid
	if req.Tag != "price 200-400" && req.Tag != "bestrated" && req.Tag != "trending" && req.Tag != "bestsellers" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tag",
		})
	}

	var items []structures.MenuItem

	// if the tag is price 200-400, run a sql query on menu items table to get all items with price between 200 and 400 and return the response direclty.
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

	if err := s.Db.Where("tag = ?", req.Tag).Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No top picks found",
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
