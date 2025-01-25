package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
)

type Server struct {
	Db *gorm.DB
}

func (s *Server) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"response": "Pong",
		"status":   "Success",
	})
}
