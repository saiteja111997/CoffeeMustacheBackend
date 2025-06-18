package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) AcceptTermsAndConditions(c *fiber.Ctx) error {

	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))
	if userId == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Create a new TermsAndConditions record
	var terms structures.TermsAndConditions
	terms.UserID = userId
	terms.AcceptedOn = time.Now()
	terms.Version = "1.0" // You can change this to the current version of your terms and conditions
	if err := s.Db.Create(&terms).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to accept terms and conditions",
		})
	}

	// Update the user's status to indicate that they have accepted the terms
	var user structures.User
	if err := s.Db.Model(&user).Where("id = ?", userId).Update("terms_accepted", true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user status",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Terms and conditions accepted successfully",
		"data": fiber.Map{
			"user_id":     userId,
			"accepted_on": terms.AcceptedOn,
			"version":     terms.Version,
		},
	})
}
