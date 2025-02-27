package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
)

// PlaceOrderRequest represents the request payload
type PlaceOrderRequest struct {
	CartID      string  `json:"cart_id"`
	SessionID   string  `json:"session_id"`
	TotalAmount float64 `json:"total_amount"`
}

// PlaceOrderResponse represents the response payload
type PlaceOrderResponse struct {
	OrderID string `json:"order_id"`
}

// PlaceOrder handles order creation
func (s *Server) PlaceOrder(c *fiber.Ctx) error {
	var req PlaceOrderRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.CartID == "" || req.SessionID == "" || req.TotalAmount <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "cart_id, session_id, and total_amount are required",
		})
	}

	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))
	if userId == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Generate a new Order ID using ksuid
	orderID := ksuid.New().String()

	// Create a new order instance
	order := structures.Order{
		OrderID:       orderID,
		CartID:        req.CartID,
		SessionID:     req.SessionID,
		UserID:        userId,
		OrderStatus:   structures.OrderPlaced, // Set status to "Placed"
		PaymentStatus: structures.Pending,     // Set payment status to "Pending"
		TotalAmount:   req.TotalAmount,
		OrderTime:     time.Now(),
	}

	// Insert into the database
	if err := s.Db.Create(&order).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to place order",
		})
	}

	// Use a wait group to perform both updates in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	// Error variables to capture errors from goroutines
	var cartUpdateErr, cartItemsUpdateErr error

	// Update cart status to "Ordered"
	go func() {
		defer wg.Done()
		cartUpdateErr = s.Db.Model(&structures.Cart{}).
			Where("cart_id = ?", req.CartID).
			Update("cart_status", structures.CartOrdered).Error
	}()

	// Update all the cart items with this cart id as ordered
	go func() {
		defer wg.Done()
		cartItemsUpdateErr = s.Db.Model(&structures.CartItem{}).
			Where("cart_id = ?", req.CartID).
			Update("cart_status", structures.CartItemOrdered).Error
	}()

	// Wait for both updates to complete
	wg.Wait()

	// Check if any errors occurred
	if cartUpdateErr != nil || cartItemsUpdateErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart or cart items status",
		})
	}

	// Return the generated order ID
	return c.Status(http.StatusOK).JSON(PlaceOrderResponse{
		OrderID: orderID,
	})
}
