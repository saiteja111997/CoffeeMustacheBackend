package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type FetchCafeDetailsRequest struct {
	CafeID uint `json:"cafe_id"`
}

func (s *Server) GetCafeDetails(c *fiber.Ctx) error {

	fmt.Println("Getting cafe details... ")

	var request FetchCafeDetailsRequest

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var cafeResponse structures.Cafe

	// Fetch cafe details from your database
	err := s.Db.Where("id =?", request.CafeID).First(&cafeResponse).Error
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cafe details",
		})
	}

	// Return cafe details
	return c.JSON(fiber.Map{
		"data": cafeResponse,
	})

}
