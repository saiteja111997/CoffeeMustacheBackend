package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
)

// PlaceOrder handles order creation
func (s *Server) PlaceOrder(c *fiber.Ctx) error {
	var req structures.PlaceOrderRequest

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
	wg.Add(3)

	// Error variables to capture errors from goroutines
	var cartUpdateErr, cartItemsUpdateErr, discountUpdateErr error

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
			Update("status", structures.CartItemOrdered).Error
	}()

	// Update the discounts table.
	go func() {
		defer wg.Done()
		discountUpdateErr = s.Db.Model(&structures.Discount{}).
			Where("cart_id = ?", req.CartID).
			Updates(map[string]interface{}{
				"user_id":         userId,
				"total_cost":      req.TotalAmount,
				"order_id":        orderID,
				"discount_amount": req.Discount, // Assuming req.DiscountAmount exists
			}).Error
	}()
	wg.Wait()

	if discountUpdateErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update discount",
		})
	}

	// Check if any errors occurred
	if cartUpdateErr != nil || cartItemsUpdateErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart or cart items status",
		})
	}

	// Return the generated order ID
	return c.Status(http.StatusOK).JSON(structures.PlaceOrderResponse{
		OrderID: orderID,
	})
}

func (s *Server) FetchOrderDetails(c *fiber.Ctx) error {
	var req structures.FetchOrderDetailsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.SessionID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "session_id is required"})
	}

	// 1) Fetch all orders for the session (payment pending/failed example)
	var orders []structures.Order
	if err := s.Db.Where(
		"session_id = ? AND (payment_status = ? OR payment_status = ?)",
		req.SessionID, "Pending", "Failed",
	).Find(&orders).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}

	// This will store final results grouped by user
	results := make(structures.UserOrdersMap)

	// We'll need a mutex to safely update 'results'
	var mu sync.Mutex

	// We'll wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(len(orders))

	// 2) For each order, spin up a goroutine
	for _, ord := range orders {
		// Capture the order in the loop variable
		order := ord

		go func() {
			defer wg.Done()

			// Prepare concurrency for cart items and user details
			var (
				cartItems    []structures.CartItem
				cartItemsErr error
			)

			// Only fetch items with "Ordered" status (as an example)
			cartItemsErr = s.Db.Where(
				"cart_id = ? AND status = ?",
				order.CartID, structures.CartItemOrdered,
			).Find(&cartItems).Error

			// Handle errors if any
			if cartItemsErr != nil {
				// You might want to log or handle partial results differently
				fmt.Println("Failed to fetch user or cart items:", cartItemsErr)
				return
			}

			// Parse customizations for each cart item
			var cartItemDetails []structures.CartItemDetail
			for _, ci := range cartItems {
				// Parse JSON customizations
				var crossSell []structures.CrossSells
				var customizations []structures.Customization

				// See is customisations exist
				if len(ci.CustomizationIDs) == 0 {
					fmt.Println("No customizations found")
				} else {

					var customizationIDs []string
					if err := json.Unmarshal(ci.CustomizationIDs, &customizationIDs); err != nil {
						fmt.Println("Failed to unmarshal customization IDs:", err)
						continue
					}

					// Run a for loop on customizations and fetch the id and name from item customizations table
					for _, id := range customizationIDs {
						var customization structures.ItemCustomization
						if err := s.Db.Where("id = ?", id).First(&customization).Error; err != nil {
							fmt.Println("Failed to fetch customization:", err)
							continue
						}
						customizations = append(customizations, structures.Customization{
							ID:         customization.ID,
							OptionName: customization.OptionName,
						})
					}

					fmt.Println("Printin customization", customizations)
				}

				if len(ci.CrossSellItemIDs) == 0 {
					fmt.Println("No cross sells found")
				} else {
					var crossSellIDs []string
					if err := json.Unmarshal(ci.CrossSellItemIDs, &crossSellIDs); err != nil {
						fmt.Println("Failed to unmarshal cross sell IDs:", err)
						continue
					}

					// Run a for loop on cross sell ids to fetch the name and id of the items from menu items table
					for _, id := range crossSellIDs {
						var item structures.MenuItem
						if err := s.Db.Where("id =?", id).First(&item).Error; err != nil {
							fmt.Println("Failed to fetch item:", err)
							continue
						}
						crossSell = append(crossSell, structures.CrossSells{
							ID:   item.ID,
							Name: item.Name,
						})
					}
				}

				// Find item name based on the item id from menu items table
				var item structures.MenuItem
				if err := s.Db.Where("id = ?", ci.ItemID).First(&item).Error; err != nil {
					fmt.Println("Failed to fetch item:", err)
					continue
				}

				cartItemDetails = append(cartItemDetails, structures.CartItemDetail{
					ItemName:       item.Name,
					CartItemID:     ci.CartItemID,
					ItemID:         ci.ItemID,
					Quantity:       ci.Quantity,
					Price:          ci.Price,
					SpecialRequest: ci.SpecialRequest,
					Customizations: customizations,
					CrossSells:     crossSell,
				})
			}

			// Build an OrderResponse
			response := structures.OrderResponse{
				OrderID:     order.OrderID,
				CartID:      order.CartID,
				CartItems:   cartItemDetails,
				TotalAmount: order.TotalAmount,
			}

			// 3) Concurrency-safe append to 'results'
			mu.Lock()
			results[order.UserID] = append(results[order.UserID], response)
			mu.Unlock()
		}()
	}

	// 4) Wait for all fetch goroutines to complete
	wg.Wait()

	// 5) Transform the map into your final JSON shape
	// e.g. you might want an array of objects:
	// [{ user_name: 'Alice', orders: [...] }, ...]

	var finalResponse = make(map[string]structures.FinalResponse)

	for userID, userOrders := range results {
		if len(userOrders) == 0 {
			continue
		}

		// Fetch user details from the database
		var user structures.User
		if err := s.Db.Where("id = ?", userID).First(&user).Error; err != nil {
			fmt.Println("Failed to fetch user details:", err)
			continue
		}

		var totalAmount float64

		for _, userOrder := range userOrders {
			totalAmount += userOrder.TotalAmount
		}

		finalResponse[user.Name] = structures.FinalResponse{
			UserID:               userID,
			CumilativeOrderTotal: totalAmount,
			Orders:               userOrders,
		}
	}

	// 6) Return JSON
	return c.Status(http.StatusOK).JSON(finalResponse)
}
