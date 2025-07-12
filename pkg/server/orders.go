package server

import (
	"coffeeMustacheBackend/pkg/helper"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

// PlaceOrder handles order creation
func (s *Server) PlaceOrder(c *fiber.Ctx) error {
	var req structures.PlaceOrderRequest

	fmt.Println("Placing Order ...")

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.CartID == "" || req.SessionID == "" || req.TotalAmount <= 0 {
		fmt.Println("Missing required fields")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "cart_id, session_id, and total_amount are required",
		})
	}

	fmt.Println("Printing Cart Id and Session Id", req.CartID, req.SessionID)

	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))
	if userId == 0 {
		fmt.Println("User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get the status of the session using session ID
	var session structures.Session
	if err := s.Db.Where("session_id = ?", req.SessionID).First(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}
	sessionStatus := session.SessionStatus

	// Check if the session is inactive
	if structures.SessionStatus(sessionStatus) != structures.Active {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session is inactive",
		})
	}

	// Generate a new Order ID using ksuid
	orderID := ksuid.New().String()
	fmt.Println("Printing Order ID", orderID)

	// Create a new order instance
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Failed to load location:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process order time",
		})
	}

	order := structures.Order{
		OrderID:        orderID,
		CafeId:         req.CafeID,
		CartID:         req.CartID,
		SessionID:      req.SessionID,
		UserID:         userId,
		SpecialRequest: req.SpecialRequest,
		OrderStatus:    structures.OrderPlaced, // Set status to "Placed"
		PaymentStatus:  structures.Pending,     // Set payment status to "Pending"
		TotalAmount:    req.TotalAmount,
		OrderTime:      time.Now().In(location).Truncate(time.Second), // Use Asia/Kolkata timezone
	}

	// Insert into the database
	if err := s.Db.Create(&order).Error; err != nil {
		fmt.Println("Error placing order:", err)
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
			Update("cart_status", string(structures.CartOrdered)).Error

		fmt.Println("Cart status updated to ordered:", cartUpdateErr)

	}()

	// Update all the cart items with this cart id as ordered
	go func() {
		defer wg.Done()
		cartItemsUpdateErr = s.Db.Model(&structures.CartItem{}).
			Where("cart_id = ? AND status NOT IN ('Canceled')", req.CartID).
			Update("status", structures.CartItemOrdered).Error

		fmt.Println("Cart items status updated to ordered:", cartItemsUpdateErr)

	}()

	// Insert into the discounts table.
	go func() {
		defer wg.Done()
		discount := structures.Discount{
			UserId:        userId,
			CafeID:        req.CafeID,   // Assuming req.CafeID exists
			DiscountValue: req.Discount, // Assuming req.DiscountValue exists
			TotalCost:     req.TotalAmount,
			OrderId:       orderID,
		}
		discountUpdateErr = s.Db.Create(&discount).Error

		fmt.Println("Discount inserted:", discountUpdateErr)

	}()
	wg.Wait()

	if discountUpdateErr != nil {
		fmt.Println("Error updating discount:", discountUpdateErr)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update discount",
		})
	}

	// Check if any errors occurred
	if cartUpdateErr != nil || cartItemsUpdateErr != nil {
		fmt.Println("Error updating cart or cart items:", cartUpdateErr, cartItemsUpdateErr)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart or cart items status",
		})
	}

	// Get the cart total from cart table based on cart id
	var cart structures.Cart
	if err := s.Db.Where("cart_id = ?", req.CartID).First(&cart).Error; err != nil {
		fmt.Println("Failed to fetch cart details:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cart details",
		})
	}

	var loyaltyPoints uint

	// Get upsell_data entry for the given cart ID
	var upsellData structures.UpsellData
	if err := s.Db.Where("cart_id = ?", req.CartID).First(&upsellData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("No upsell data found for the cart")
		} else {
			fmt.Println("Failed to fetch upsell data:", err)
		}
	} else {
		if cart.TotalAmount >= upsellData.TargetAmount {
			// If the cart total is greater than or equal to the target amount, add the mustaches to the loyalty points
			loyaltyPoints = upsellData.MustachesToGive
			fmt.Println("Loyalty points earned from upsell data:", loyaltyPoints)

			// Update OfferAccepted to true in upsell_data table
			if err := s.Db.Model(&structures.UpsellData{}).
				Where("cart_id = ?", req.CartID).
				Update("offer_accepted", true).Error; err != nil {
				fmt.Println("Failed to update upsell data:", err)
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update upsell data",
				})
			}
			fmt.Println("Upsell data updated successfully")

		} else {
			fmt.Println("Cart total is less than the target amount for upsell data")
			loyaltyPoints = uint(cart.TotalAmount / 50)
		}
	}

	// Update the reward_transactions table with the earned loyalty points for the user
	earnedDate := time.Now().In(location).Truncate(time.Second) // Use Asia/Kolkata timezone
	if err := s.Db.Create(&structures.RewardTransaction{
		UserID:          userId,
		CafeID:          req.CafeID,
		SessionID:       req.SessionID,
		TransactionType: "credited",
		Mustaches:       loyaltyPoints,
		EarnedDate:      &earnedDate,
	}).Error; err != nil {
		fmt.Println("Failed to update reward transactions:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update reward transactions",
		})
	}

	// Send a push notification by fetching device tokens from fcm_tokens table based on cafe id
	var deviceTokens []string
	if err := s.Db.Model(&structures.FcmToken{}).
		Where("cafe_id = ?", req.CafeID).
		Pluck("token", &deviceTokens).Error; err != nil {
		fmt.Println("Failed to fetch device tokens:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch device tokens",
		})
	}

	body := fmt.Sprintf("New order received for Table No: %s", session.TableName)

	if len(deviceTokens) > 0 {
		if err := helper.SendPushNotification(deviceTokens, "Order Update", body); err != nil {
			fmt.Println("Failed to send push notification:", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send push notification",
			})
		}
	} else {
		fmt.Println("No device tokens found for the cafe")
	}

	fmt.Println("Order placed successfully with ID:", orderID)

	// Return the generated order ID
	return c.Status(http.StatusOK).JSON(structures.PlaceOrderResponse{
		OrderID: orderID,
		Rewards: loyaltyPoints, // Return the earned loyalty points
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

	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))
	if userId == 0 {
		fmt.Println("User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	cafeID := uint(c.Locals("cafeId").(float64))

	// GEt current time in Asia/Kolkata timezone
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Failed to load location:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process time",
		})
	}

	currentTime := time.Now().In(location).Truncate(time.Second) // Use Asia/Kolkata timezone

	// Get Start time of the day in Asia/Kolkata timezone
	startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, location)

	// Get complete_pos status for the cafe
	var cafe structures.Cafe
	if err := s.Db.Where("id = ?", cafeID).First(&cafe).Error; err != nil {
		fmt.Println("Failed to fetch cafe details:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cafe details",
		})
	}

	// 1) Fetch all orders for the session (payment pending/failed)
	var orders []structures.Order

	if cafe.Complete_Pos {
		if err := s.Db.Where(
			"session_id = ? AND (payment_status = ? OR payment_status = ?) AND order_status != ?",
			req.SessionID, "Pending", "Failed", structures.OrderCancelled,
		).Find(&orders).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
		}
	} else {
		if err := s.Db.Where(
			"session_id = ? AND (payment_status = ? OR payment_status = ?) AND order_status != ? AND user_id = ? AND order_time > ?",
			req.SessionID, "Pending", "Failed", structures.OrderCancelled, userId, startOfDay,
		).Find(&orders).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
		}
	}

	fmt.Println("Length of the orders:", len(orders))

	// This will store final results grouped by user
	results := make(structures.UserOrdersMap)

	for _, ord := range orders {

		// If user id is 0 continue to next order
		if ord.UserID == 0 {
			continue
		}

		// Capture the order in the loop variable
		order := ord

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
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user or cart items",
			})
		}

		// Parse customizations for each cart item
		var cartItemDetails []structures.CartItemDetail
		for _, ci := range cartItems {
			// Parse JSON customizations
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
			})
		}

		// Fetch discount from the order id
		var discount structures.Discount
		if err := s.Db.Where("order_id = ?", order.OrderID).First(&discount).Error; err != nil {
			fmt.Println("Failed to fetch discount:", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch discount",
			})
		}

		// Build an OrderResponse
		response := structures.OrderResponse{
			OrderID:     order.OrderID,
			CartID:      order.CartID,
			CartItems:   cartItemDetails,
			OrderedAt:   order.OrderTime,
			Discount:    discount.DiscountValue,
			TotalAmount: order.TotalAmount,
		}

		results[order.UserID] = append(results[order.UserID], response)
	}

	// 5) Transform the map into your final JSON shape
	// e.g. you might want an array of objects:
	// [{ user_name: 'Alice', orders: [...] }, ...]

	var finalResponse = make(map[string]interface{})
	var finalTimeStamp time.Time

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
		var totalDiscount float64

		for _, userOrder := range userOrders {
			// calculate the latest timestamp
			if userOrder.OrderedAt.After(finalTimeStamp) {
				finalTimeStamp = userOrder.OrderedAt
			}
			totalAmount += userOrder.TotalAmount
			totalDiscount += userOrder.Discount
		}

		finalResponse[user.Name] = structures.FinalResponse{
			UserID:               userID,
			CumilativeOrderTotal: totalAmount,
			Discount:             totalDiscount,
			Orders:               userOrders,
		}
	}

	finalResponse["timestamp"] = finalTimeStamp

	// Fetch Cafe id from sessions table using session ID
	// var sessionCafe struct {
	// 	CafeID uint
	// }
	// if err := s.Db.Model(&structures.Session{}).
	// 	Select("cafe_id").
	// 	Where("session_id = ?", req.SessionID).
	// 	Scan(&sessionCafe).Error; err != nil {
	// 	fmt.Println("Failed to fetch cafe ID:", err)
	// 	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": "Failed to fetch cafe ID",
	// 	})
	// }
	// cafeID := sessionCafe.CafeID

	// Fetch if there is any advertisement for the cafe, where the current date is between start and end date
	var advertisement structures.CafeAdvertisement
	if err := s.Db.Where(
		"cafe_id = ? AND ad_start_time <= ? AND ad_end_time >= ? AND ad_status = ?",
		cafeID, currentTime, currentTime, "active",
	).Order("created_at desc").First(&advertisement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("No advertisement found for the cafe")
		} else {
			fmt.Println("Failed to fetch advertisement:", err)
		}
	} else {
		fmt.Println("Advertisement found for the cafe:", advertisement)
		finalResponse["advertisement"] = advertisement
	}

	// 6) Return JSON
	return c.Status(http.StatusOK).JSON(finalResponse)
}
