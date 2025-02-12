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

	var topPicks []structures.MenuItem
	if err := s.Db.Where("tag = ?", req.Tag).Find(&topPicks).Error; err != nil {
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
		"top_picks": topPicks,
	})
}
