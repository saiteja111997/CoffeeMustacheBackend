package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) CallWaiter(c *fiber.Ctx) error {

	fmt.Println("Calling Waiter ...")

	userID := uint(c.Locals("userId").(float64))

	var request structures.CallWaiterRequest

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var customerRequest structures.CustomerRequest

	customerRequest.SessionID = request.SessionID
	customerRequest.TableNumber = request.TableNumber
	customerRequest.RequestType = request.RequestType
	customerRequest.UserID = userID
	customerRequest.CafeID = request.CafeID

	// Get time in Asia/Kolkata zone
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to load location",
		})
	}
	currentTime := time.Now().In(location)

	customerRequest.RequestedAt = currentTime

	// Save customer request to database
	if err := s.Db.Create(&customerRequest).Error; err != nil {
		fmt.Println("Error saving customer request:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save customer request",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Customer request received successfully",
	})
}
