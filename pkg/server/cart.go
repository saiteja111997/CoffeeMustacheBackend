package server

import (
	"coffeeMustacheBackend/pkg/helper"
	"coffeeMustacheBackend/pkg/structures"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (s *Server) AddToCart(c *fiber.Ctx) error {
	var req structures.AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))

	// Validate input
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one item must be provided",
		})
	}

	// If the session is inactive, return an error
	if req.SessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
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

	// Initialize cart ID
	cartID := req.CartID

	// Check if Cart ID is provided
	if cartID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cart ID is required",
		})
	}

	// Check if the cart already exists in the database
	var existingCart structures.Cart
	result := s.Db.Where("cart_id = ?", cartID).First(&existingCart)
	if result.RowsAffected == 0 {
		// Cart does not exist, create a new one
		newCart := structures.Cart{
			CartID:         cartID,
			CafeId:         req.CafeID,
			SessionID:      req.SessionID,
			UserID:         userId,
			CartStatus:     structures.CartActive,
			TotalAmount:    req.TotalAmount,
			DiscountAmount: req.DiscountAmount,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Insert new cart into the database
		if err := s.Db.Create(&newCart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create cart",
			})
		}
	} else if result.Error != nil {
		// Database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	} else {
		// Cart exists, update the amounts
		if err := s.Db.Model(&structures.Cart{}).
			Where("cart_id = ?", cartID).
			Updates(map[string]interface{}{
				"total_amount":    req.TotalAmount,
				"discount_amount": req.DiscountAmount,
				"updated_at":      time.Now(),
			}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cart amounts",
			})
		}
	}

	// Add multiple items to cart
	for _, item := range req.Items {

		addedVia := structures.CartInsertType(item.AddedVia)

		// Check if added_via is valid
		if !helper.IsValid(addedVia) {
			fmt.Println("Invalid added_via: ", addedVia)
		}

		newCartItem := structures.CartItem{
			CartItemID:       item.CartItemId,
			CartID:           cartID,
			CafeID:           req.CafeID,
			ItemID:           item.ItemID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			AddedAt:          time.Now(),
			AddedVia:         addedVia,
			SpecialRequest:   item.SpecialRequest,
			CrossSellItemIDs: item.CrossSellItemIDs,
			CustomizationIDs: item.CustomizationIDs,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// Get category id from the menu items table.
		var menuItem structures.MenuItem
		if err := s.Db.Where("id = ?", item.ItemID).First(&menuItem).
			Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Menu item not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}

		// For the given item id, increment the category counter based on category id
		if menuItem.CategoryID != 0 {
			var category structures.Category
			if err := s.Db.Where("id = ?", menuItem.CategoryID).First(&category).
				Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "Category not found",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Database error",
				})
			}
			category.Counter += 1
			if err := s.Db.Model(&structures.Category{}).
				Where("id = ?", category.ID).
				Update("counter", category.Counter).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update category counter",
				})
			}
		}

		// If added via is UpgradeCartAi, get the latest upgrade data for the give cart id and update the user action field to "added"
		if addedVia == structures.UpgradeCartAi {
			var latestUpgrade structures.UpdateCartResult
			if err := s.Db.Where("cart_id = ? AND suggested_item_id = ?", cartID, item.ItemID).
				Order("created_at DESC").First(&latestUpgrade).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "No upgrade data found for the given cart ID and item ID",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Database error",
				})
			}

			// Update the user action to "added"
			if err := s.Db.Model(&structures.UpdateCartResult{}).
				Where("id = ?", latestUpgrade.ID).
				Update("user_action", "added").Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update upgrade data",
				})
			}
		}

		// Insert the new cart item into the database
		if err := s.Db.Create(&newCartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add item to cart",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Items added to cart successfully",
		"cart_id": cartID,
	})
}

func (s *Server) GetCart(c *fiber.Ctx) error {
	// Fetch user ID from c.Locals
	userID, ok := c.Locals("userId").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized request",
		})
	}

	// Parse request body
	var req structures.GetCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if cart exists and belongs to the user, session
	var cart structures.Cart
	if err := s.Db.Where("cart_id = ? AND session_id = ? AND user_id = ?", req.CartID, req.SessionID, uint(userID)).
		First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Cart not found or does not belong to user",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// Check if cart is active
	if cart.CartStatus != structures.CartActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cart is not active",
		})
	}

	// Fetch cart items
	var cartItems []structures.CartItem
	if err := s.Db.Where("cart_id = ? AND status != ?", req.CartID, "Canceled").Find(&cartItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch cart items",
		})
	}

	// Prepare response with cart items
	var cartItemResponses []structures.CartItemResponse
	for _, item := range cartItems {
		// Convert JSONB fields to string arrays
		var customizationIDs, crossSellItemIDs []string

		// Unmarshal the JSONB data into string arrays
		if err := json.Unmarshal(item.CustomizationIDs, &customizationIDs); err != nil {
			customizationIDs = []string{} // Default to empty array in case of error
		}

		if err := json.Unmarshal(item.CrossSellItemIDs, &crossSellItemIDs); err != nil {
			crossSellItemIDs = []string{} // Default to empty array in case of error
		}

		// Fetch customization names from the DB
		var customizations []structures.ItemCustomization
		if len(customizationIDs) > 0 {
			if err := s.Db.Where("id IN (?)", customizationIDs).Find(&customizations).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to fetch customization details",
				})
			}
		}

		// Prepare customizations response with both ID & Name
		var customizationDetails []map[string]string
		for _, customization := range customizations {
			customizationDetails = append(customizationDetails, map[string]string{
				"id":   fmt.Sprintf("%d", customization.ID),
				"name": customization.OptionName,
			})
		}

		cartItemResponses = append(cartItemResponses, structures.CartItemResponse{
			CartItemId:           item.CartItemID,
			ItemID:               item.ItemID,
			Quantity:             item.Quantity,
			Price:                item.Price,
			AddedVia:             string(item.AddedVia),
			SpecialRequest:       item.SpecialRequest,
			CustomizationDetails: customizationDetails, // Includes both ID & Name
			CrossSellItemIDs:     crossSellItemIDs,
		})
	}

	// Return cart details with items
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Cart retrieved successfully",
		"cart_id":    cart.CartID,
		"session_id": cart.SessionID,
		"user_id":    cart.UserID,
		"items":      cartItemResponses,
	})
}

func (s *Server) UpdateCustomizations(c *fiber.Ctx) error {
	var req structures.UpdateCustomizationsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if cart item exists
	var cartItem structures.CartItem
	if err := s.Db.Where("cart_item_id = ?", req.CartItemId).
		First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Cart item not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// Convert to JSON
	customizationJSON, _ := json.Marshal(req.CustomizationIDs)

	// Update customizations using cart ID and item ID
	if err := s.Db.Model(&structures.CartItem{}).
		Where("cart_item_id = ?", req.CartItemId).
		Updates(map[string]interface{}{
			"customization_ids": customizationJSON,
			"updated_at":        time.Now(),
			"price":             req.Price,
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update customizations",
		})
	}

	// Update total cart amount by direct reading CartAmount from the request body
	if err := s.Db.Model(&structures.Cart{}).
		Where("cart_id = ?", cartItem.CartID).
		Update("total_amount", req.CartAmount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart total amount",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Customizations updated successfully",
	})
}

func (s *Server) UpdateCrossSellItems(c *fiber.Ctx) error {
	var req structures.UpdateCrossSellItemsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if cart item exists
	var cartItem structures.CartItem
	if err := s.Db.Where("cart_item_id = ?", req.CartItemId).
		First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Cart item not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	// Convert to JSON
	crossSellJSON, _ := json.Marshal(req.CrossSellItemIDs)

	// Update cross-sell items using cart ID and item ID
	if err := s.Db.Model(&structures.CartItem{}).
		Where("cart_item_id = ?", req.CartItemId).
		Updates(map[string]interface{}{
			"cross_sell_item_ids": crossSellJSON,
			"updated_at":          time.Now(),
			"price":               req.Price,
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cross-sell items",
		})
	}

	// Update total cart amount by direct reading CartAmount from the request body
	if err := s.Db.Model(&structures.Cart{}).
		Where("cart_id = ?", cartItem.CartID).
		Update("total_amount", req.CartAmount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart total amount",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cross-sell items updated successfully",
	})
}

func (s *Server) UpdateQuantity(c *fiber.Ctx) error {
	var req structures.UpdateQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if cart item exists
	var cartItem structures.CartItem
	if err := s.Db.Where("cart_item_id = ?", req.CartItemId).
		First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Cart item not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if cartItem.Status == "Canceled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot update quantity of a canceled item",
		})
	}

	// If quantity is zero, mark item as "Canceled"
	if req.Quantity == 0 {
		fmt.Println("Quantity is Zero, marking item as Canceled")

		fmt.Println("Printing Status before updating : ", string(structures.CartItemCanceled))
		if err := s.Db.Model(&structures.CartItem{}).
			Where("cart_item_id = ?", req.CartItemId).
			Updates(map[string]interface{}{
				"status":     "Canceled",
				"updated_at": time.Now(),
				"quantity":   0,
			}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to cancel cart item",
			})
		}

		// Update total cart amount by directly reading CartAmount from the request body
		if err := s.Db.Model(&structures.Cart{}).
			Where("cart_id = ?", cartItem.CartID).
			Update("total_amount", req.CartAmount).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cart total amount",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Cart item marked as canceled",
		})
	}

	// Update the quantity
	if err := s.Db.Model(&structures.CartItem{}).
		Where("cart_item_id = ?", req.CartItemId).
		Updates(map[string]interface{}{
			"quantity":   req.Quantity,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update quantity",
		})
	}

	// Update total cart amount by directly reading CartAmount from the request body
	if err := s.Db.Model(&structures.Cart{}).
		Where("cart_id = ?", cartItem.CartID).
		Update("total_amount", req.CartAmount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart total amount",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Quantity updated successfully",
		"quantity": req.Quantity,
	})
}
