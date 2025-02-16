package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AddToCartRequest struct {
	CartID           string         `json:"cart_id"`
	SessionID        string         `json:"session_id" validate:"required"`
	UserID           uint           `json:"user_id" validate:"required"`
	ItemID           uint           `json:"item_id" validate:"required"`
	Quantity         int            `json:"quantity" validate:"required,min=1"`
	Price            float64        `json:"price" validate:"required"`
	SpecialRequest   string         `json:"special_request"`
	CustomizationIDs datatypes.JSON `json:"customization_ids"`   // Expecting JSON array like ["1", "2", "3"]
	CrossSellItemIDs datatypes.JSON `json:"cross_sell_item_ids"` // Expecting JSON array like ["1", "2", "3"]
}

func (s *Server) AddToCart(c *fiber.Ctx) error {
	// Define the request structure

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Initialize the cart ID
	cartID := req.CartID
	totalAmount := req.Price * float64(req.Quantity)

	// Check if Cart ID is provided
	if cartID == "" {
		// Create a new Cart with a unique CartID using ksuid
		cartID = ksuid.New().String()
		newCart := structures.Cart{
			CartID:      cartID,
			SessionID:   req.SessionID,
			UserID:      req.UserID,
			CartStatus:  structures.CartActive,
			TotalAmount: totalAmount,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Insert new cart into the database
		if err := s.Db.Create(&newCart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create cart",
			})
		}
	} else {
		// If Cart ID exists, update the total amount
		if err := s.Db.Model(&structures.Cart{}).
			Where("cart_id = ?", cartID).
			Update("total_amount", gorm.Expr("total_amount + ?", totalAmount)).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cart total amount",
			})
		}
	}

	// Create a new Cart Item
	newCartItem := structures.CartItem{
		CartID:           cartID,
		ItemID:           req.ItemID,
		Quantity:         req.Quantity,
		Price:            req.Price,
		AddedAt:          time.Now(),
		SpecialRequest:   req.SpecialRequest,
		CrossSellItemIDs: req.CrossSellItemIDs,
		CustomizationIDs: req.CustomizationIDs,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Insert the new cart item into the database
	if err := s.Db.Create(&newCartItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add item to cart",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Item added to cart successfully",
		"cart_id":      cartID,
		"total_amount": totalAmount,
	})
}
