package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetPersonalisedData(c *fiber.Ctx) error {

	userId := uint(c.Locals("userId").(float64))
	cafeId := uint(c.Locals("cafeId").(float64))

	type ItemObject = structures.ItemObject

	// Final response structure
	result := fiber.Map{
		"repeat_order": []ItemObject{},
		"recent_order": []ItemObject{},
		"favourites":   []ItemObject{},
	}

	// STEP 1: Repeat Orders
	type OrderGroup struct {
		OrderTime time.Time
		ItemIDs   []uint
	}

	rows, err := s.Db.Raw(`
		SELECT o.order_time, ci.item_id, o.order_id
		FROM orders o
		JOIN cart_items ci ON o.cart_id = ci.cart_id
		WHERE o.user_id = ? AND o.cafe_id = ? AND o.order_status = 'Placed' AND ci.status = 'Ordered'
		ORDER BY o.order_time DESC
	`, userId, cafeId).Rows()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error while fetching orders"})
	}
	defer rows.Close()

	orderItems := make(map[string]OrderGroup)
	for rows.Next() {
		var orderTime time.Time
		var itemId uint
		var orderId string
		if err := rows.Scan(&orderTime, &itemId, &orderId); err != nil {
			continue
		}
		group := orderItems[orderId]
		group.OrderTime = orderTime
		group.ItemIDs = append(group.ItemIDs, itemId)
		orderItems[orderId] = group
	}

	type itemSetKey struct {
		SortedIDs string
	}
	setFreq := make(map[itemSetKey]int)
	setLatestTime := make(map[itemSetKey]time.Time)
	setToIDs := make(map[itemSetKey][]uint)

	for _, group := range orderItems {
		ids := group.ItemIDs
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		key := itemSetKey{SortedIDs: fmt.Sprint(ids)}
		setFreq[key]++
		setLatestTime[key] = group.OrderTime
		setToIDs[key] = ids
	}

	var repeatSet itemSetKey
	var latest time.Time
	for k, count := range setFreq {
		if count > 1 && setLatestTime[k].After(latest) {
			latest = setLatestTime[k]
			repeatSet = k
		}
	}

	if ids, ok := setToIDs[repeatSet]; ok && len(ids) > 0 {
		var repeatItems []ItemObject
		err := s.Db.Table("menu_items").
			Select("id, cafe_id, image_url, name, price, is_customizable").
			Where("id IN (?)", ids).
			Scan(&repeatItems).Error
		if err == nil && len(repeatItems) > 0 {
			result["repeat_order"] = repeatItems
		}
	}

	// STEP 2: Recent Order
	var recentCartID string
	err = s.Db.Raw(`
		SELECT cart_id FROM orders
		WHERE user_id = ? AND cafe_id = ? AND order_status = 'Placed'
		ORDER BY order_time DESC
		LIMIT 1
	`, userId, cafeId).Row().Scan(&recentCartID)

	if err == nil && recentCartID != "" {
		var recentItems []ItemObject
		s.Db.Raw(`
			SELECT mi.id, mi.cafe_id, mi.image_url, mi.name, mi.price, mi.is_customizable
			FROM cart_items ci
			JOIN menu_items mi ON ci.item_id = mi.id
			WHERE ci.cart_id = ? AND ci.status = 'Ordered'
		`, recentCartID).Scan(&recentItems)

		if len(recentItems) > 0 {
			result["recent_order"] = recentItems
		}
	}

	// STEP 3: Favorites
	var favoriteItems []ItemObject
	s.Db.Raw(`
		SELECT mi.id, mi.cafe_id, mi.image_url, mi.name, mi.price, mi.is_customizable
		FROM item_favorites f
		JOIN menu_items mi ON f.item_id = mi.id
		WHERE f.user_id = ? AND f.cafe_id = ?
		ORDER BY f.created_at DESC
		LIMIT 3
	`, userId, cafeId).Scan(&favoriteItems)

	if len(favoriteItems) > 0 {
		result["favourites"] = favoriteItems
	}

	// Final structured response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}
