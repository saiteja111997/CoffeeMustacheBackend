package server

import (
	"coffeeMustacheBackend/pkg/helper"
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

	// Send a push notification by fetching device tokens from fcm_tokens table based on cafe id
	var deviceTokens []string
	if err := s.Db.Model(&structures.FcmToken{}).
		Where("cafe_id = ?", request.CafeID).
		Pluck("token", &deviceTokens).Error; err != nil {
		fmt.Println("Failed to fetch device tokens:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch device tokens",
		})
	}

	body := fmt.Sprintf("%s request from Table No: %s", request.RequestType, request.TableNumber)

	if len(deviceTokens) > 0 {
		if err := helper.SendPushNotification(deviceTokens, "Customer Request", body); err != nil {
			fmt.Println("Failed to send push notification:", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send push notification",
			})
		}
	} else {
		fmt.Println("No device tokens found for the cafe")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Customer request received successfully",
	})
}
