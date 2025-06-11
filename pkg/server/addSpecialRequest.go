package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AddItemCartRequest struct {
	CartItemId     string `json:"cart_item_id"`
	SpecialRequest string `json:"special_request"`
}

func (s *Server) AddSpecialRequest(c *fiber.Ctx) error {

	var req AddItemCartRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if the current cart item id exists in cart items table
	var cartItem structures.CartItem
	if err := s.Db.Where("cart_item_id =?", req.CartItemId).First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart item not found",
		})
	}

	// Update the special request for the specific cart item
	cartItem.SpecialRequest = req.SpecialRequest
	if err := s.Db.Model(&cartItem).Where("cart_item_id = ?", req.CartItemId).Update("special_request", req.SpecialRequest).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Special request updated successfully",
		"data":    cartItem,
	})

}
