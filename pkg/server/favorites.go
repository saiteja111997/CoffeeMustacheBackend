package server

import (
	"coffeeMustacheBackend/pkg/structures"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetFavouriteItems(c *fiber.Ctx) error {
	// Get user id from JWT claims
	userId := uint(c.Locals("userId").(float64))

	// Define response structure
	type FavoriteItemResponse struct {
		FavouriteID uint    `json:"favourite_id"`
		ID          uint    `json:"id"`
		CafeID      uint    `json:"cafe_id"`
		ImageURL    string  `json:"image_url"`
		Name        string  `json:"name"`
		Price       float64 `json:"price"`
	}

	var favoriteItems []FavoriteItemResponse

	// Fetch favorites joined with menu_items, ordered by latest
	err := s.Db.Table("item_favorites").
		Select("item_favorites.id AS favourite_id, item_favorites.item_id AS id, item_favorites.cafe_id, menu_items.image_url, menu_items.name, menu_items.price").
		Joins("JOIN menu_items ON item_favorites.item_id = menu_items.id").
		Where("item_favorites.user_id = ?", userId).
		Order("item_favorites.created_at DESC").
		Scan(&favoriteItems).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch favorite items",
		})
	}

	// Show most recent 3 favorite unique items
	uniqueItems := []FavoriteItemResponse{}
	seen := make(map[uint]bool)

	for _, item := range favoriteItems {
		if !seen[item.ID] {
			uniqueItems = append(uniqueItems, item)
			seen[item.ID] = true
		}
		if len(uniqueItems) == 3 {
			break
		}
	}

	// Return the favorite items
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   uniqueItems,
	})
}

func (s *Server) AddFavouriteItem(c *fiber.Ctx) error {
	// Get user id from JWT claims
	userId := uint(c.Locals("userId").(float64))
	cafeId := uint(c.Locals("cafeId").(float64))

	type AddFavoriteItemRequest struct {
		ItemID uint `json:"item_id"`
	}

	// Parse request body
	var req AddFavoriteItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate input
	if req.ItemID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "item_id and cafe_id are required",
		})
	}

	// Create favorite item
	favoriteItem := structures.ItemFavorite{
		UserID: userId,
		ItemID: req.ItemID,
		CafeID: cafeId,
	}

	if err := s.Db.Create(&favoriteItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add favorite item",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Favorite item added successfully",
		"data":    favoriteItem,
	})
}
