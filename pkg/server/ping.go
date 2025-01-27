package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
)

type Server struct {
	Db     *gorm.DB
	Config structures.Config
}

func (s *Server) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"response": "Pong",
		"status":   "Success",
	})
}
