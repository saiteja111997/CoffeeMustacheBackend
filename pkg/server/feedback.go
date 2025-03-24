package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type FeedbackFormRequest struct {
	SessionID string `json:"session_id"`
}

type FeedbackFormResponse struct {
	Name             string `json:"name"`
	ItemID           uint   `json:"item_id"`
	Image            string `json:"image"`
	ShortDescription string `json:"short_description"`
}

type FeedbackRequest struct {
	SessionID   string             `json:"session_id"`
	CafeID      uint               `json:"cafe_id"`
	Items       []ItemFeedbackData `json:"items"`       // List of items being rated
	CafeRating  int                `json:"cafe_rating"` // Overall cafe rating (1-5)
	WouldReturn int                `json:"would_return"`
	Review      string             `json:"review"` // Optional user review
}

type ItemFeedbackData struct {
	ItemID uint `json:"item_id"`
	Rating int  `json:"rating"` // Rating for each item (1-5)
}

func (s *Server) GetFeedbackForm(c *fiber.Ctx) error {
	var feedbackFormRequest FeedbackFormRequest
	if err := c.BodyParser(&feedbackFormRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Get user ID from locals
	userId, ok := c.Locals("userId").(float64)
	if !ok || userId == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Step 1: Get Cart IDs where OrderStatus is "placed"
	var cartIDs []string
	if err := s.Db.Model(&structures.Order{}).
		Where("session_id = ? AND user_id = ? AND order_status = ?", feedbackFormRequest.SessionID, uint(userId), "Placed").
		Pluck("cart_id", &cartIDs).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cart IDs",
		})
	}

	if len(cartIDs) == 0 {
		return c.JSON(fiber.Map{
			"message": "No carts found for the given session and user.",
			"items":   []structures.CartItem{},
		})
	}

	// Step 2: Fetch unique Cart Items where Status is "Ordered" using pq.Array(cartIDs)
	var cartItems []structures.CartItem
	query := `
		SELECT DISTINCT item_id
		FROM cart_items
		WHERE cart_id = ANY($1) AND status = $2
	`
	if err := s.Db.Raw(query, pq.Array(cartIDs), "Ordered").Scan(&cartItems).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve cart items",
		})
	}

	// Step 3: Fetch item names from Menu Items table
	var finalResponse []FeedbackFormResponse
	for _, item := range cartItems {
		var menuItem structures.MenuItem
		if err := s.Db.Model(&structures.MenuItem{}).
			Where("id = ?", item.ItemID).
			First(&menuItem).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve menu item",
			})
		}

		finalResponse = append(finalResponse, FeedbackFormResponse{
			Name:             menuItem.Name,
			ItemID:           item.ItemID,
			Image:            menuItem.ImageURL,
			ShortDescription: menuItem.ShortDescription,
		})
	}

	return c.JSON(fiber.Map{
		"items": finalResponse,
	})
}

func (s *Server) SubmitFeedback(c *fiber.Ctx) error {
	var feedbackReq FeedbackRequest
	if err := c.BodyParser(&feedbackReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Get user ID from locals (ensure authentication)
	userId, ok := c.Locals("userId").(float64)
	if !ok || userId == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Step 1: Insert ratings for each food item
	for _, item := range feedbackReq.Items {
		itemFeedback := structures.ItemFeedback{
			UserID:    uint(userId),
			ItemID:    item.ItemID,
			SessionID: feedbackReq.SessionID,
			Rating:    item.Rating,
			CreatedAt: time.Now(),
		}

		if err := s.Db.Create(&itemFeedback).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save item feedback",
			})
		}

		// select current rating and number of ratings from menu items table
		var menuItem structures.MenuItem
		if err := s.Db.Model(&structures.MenuItem{}).
			Where("id = ?", item.ItemID).
			First(&menuItem).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve menu item",
			})
		}

		newRating := (menuItem.Rating*float64(menuItem.TotalRatings) + float64(item.Rating)) / float64(menuItem.TotalRatings+1)

		// Update the menu items table with new rating and total ratings
		if err := s.Db.Model(&structures.MenuItem{}).
			Where("id = ?", item.ItemID).
			Updates(map[string]interface{}{
				"rating":        newRating,
				"total_ratings": menuItem.TotalRatings + 1,
			}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update menu item",
			})
		}

	}

	// Step 2: Insert overall cafe feedback
	cafeFeedback := structures.CafeFeedback{
		UserID:      uint(userId),
		CafeID:      feedbackReq.CafeID,
		Rating:      feedbackReq.CafeRating,
		SessionID:   feedbackReq.SessionID,
		WouldReturn: feedbackReq.WouldReturn,
		Review:      feedbackReq.Review,
		CreatedAt:   time.Now(),
	}

	if feedbackReq.CafeRating == 0 && len(feedbackReq.Review) == 0 {
		fmt.Println("Skipping feedback for cafe")
	} else {
		if err := s.Db.Create(&cafeFeedback).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save cafe feedback",
			})
		}

		// select current rating and number of ratings from cafes table
		var cafe structures.Cafe
		if err := s.Db.Model(&structures.Cafe{}).
			Where("id = ?", feedbackReq.CafeID).
			First(&cafe).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve cafe",
			})
		}

		newCafeRating := (cafe.Rating*float64(cafe.TotalRatings) + float64(cafe.Rating)) / float64(cafe.TotalRatings+1)

		// Update the cafes table with new rating and total ratings
		if err := s.Db.Model(&structures.Cafe{}).
			Where("id = ?", feedbackReq.CafeID).
			Updates(map[string]interface{}{
				"rating":        newCafeRating,
				"total_ratings": cafe.TotalRatings + 1,
			}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cafe",
			})
		}

	}

	// Return success response
	return c.JSON(fiber.Map{
		"message": "Feedback submitted successfully",
	})
}
