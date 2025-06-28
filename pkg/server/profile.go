package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type UserProfileResponse struct {
	Name                    string `json:"name"`
	Phone                   string `json:"phone"`
	JoinedAt                string `json:"joined_at"`
	CurrentLevel            string `json:"current_level"`
	NextLevel               string `json:"next_level"`
	DueMustacheForNextLevel uint   `json:"due_mustache_for_next_level"`
	MustacheEarnedLastMonth uint   `json:"mustache_earned_last_month"`
	CurrentBalance          uint   `json:"current_balance"`
	TotalOrders             uint   `json:"total_orders"`
}

func (s *Server) GetProfile(c *fiber.Ctx) error {

	userID := uint(c.Locals("userId").(float64))

	var user structures.User

	// Fetch user details from the database
	if err := s.Db.First(&user, userID).Error; err != nil {
		fmt.Println("Error fetching user profile:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch user profile",
		})
	}

	// Get total number of mustache from reward_transactions table where transaction_type is 'credited'
	var totalMustache int64
	if err := s.Db.Model(&structures.RewardTransaction{}).
		Select("COALESCE(SUM(mustaches), 0)").
		Where("user_id = ? AND transaction_type = ?", userID, "credited").
		Row().Scan(&totalMustache); err != nil {
		fmt.Println("Error fetching total mustache:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch total mustache",
		})
	}

	// Get total number of orders from orders table where order_status is 'Paid'
	var totalOrders int64
	if err := s.Db.Model(&structures.Order{}).
		Where("user_id = ? AND payment_status = ?", userID, "paid").
		Count(&totalOrders).Error; err != nil {
		fmt.Println("Error fetching total orders:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch total orders",
		})
	}

	// Convert int64 counts to uint for response struct
	totalMustacheUint := uint(totalMustache)
	totalOrdersUint := uint(totalOrders)

	// 🧋 Dripstarter	0–99	Welcome to the club
	// 🫖 Brew Buddy	100–299	You're warming up
	// ☕ Bean Boss	300–699	+10% bonus Mustaches per order
	// 🥇 Caffeine Royalty	700+	+15% bonus, early access to deals
	var currentLevel string
	var nextLevel string
	var dueMustacheForNextLevel uint
	var mustacheEarnedThisMonth uint

	if totalMustache < 100 {
		currentLevel = "🧋 Dripstarter"
		nextLevel = "🫖 Brew Buddy"
		dueMustacheForNextLevel = 100 - totalMustacheUint
	}
	if totalMustache >= 100 && totalMustache < 300 {
		currentLevel = "🫖 Brew Buddy"
		nextLevel = "☕ Bean Boss"
		dueMustacheForNextLevel = 300 - totalMustacheUint
	}
	if totalMustache >= 300 && totalMustache < 700 {
		currentLevel = "☕ Bean Boss"
		nextLevel = "🥇 Caffeine Royalty"
		dueMustacheForNextLevel = 700 - totalMustacheUint
	}
	if totalMustache >= 700 {
		currentLevel = "🥇 Caffeine Royalty"
		nextLevel = "No more levels"
		dueMustacheForNextLevel = 0 // No more levels after this
	}

	// Get mustache earned from the start of this month
	startOfMonth := fmt.Sprintf("%d-%02d-01", user.CreatedAt.Year(), user.CreatedAt.Month())
	var mustacheEarnedThisMonthInt int64 = 0
	if err := s.Db.Model(&structures.RewardTransaction{}).
		Select("COALESCE(SUM(mustaches), 0)").
		Where("user_id = ? AND transaction_type = ? AND created_at >= ?", userID, "credited", startOfMonth).
		Row().Scan(&mustacheEarnedThisMonthInt); err != nil {
		fmt.Println("Error fetching mustache earned this month:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch mustache earned this month",
		})
	}
	mustacheEarnedThisMonth = uint(mustacheEarnedThisMonthInt)

	// Convert user's created_at to a string for JSON response Eg 5th May
	joinedAtStr := user.CreatedAt.Format("2 Jan 2006")
	if user.CreatedAt.IsZero() {
		joinedAtStr = "Unknown"
	}
	// Return the user profile details
	return c.JSON(UserProfileResponse{
		Name:                    user.Name,
		Phone:                   user.Phone,
		JoinedAt:                joinedAtStr,
		CurrentLevel:            currentLevel,
		NextLevel:               nextLevel,
		DueMustacheForNextLevel: dueMustacheForNextLevel,
		MustacheEarnedLastMonth: mustacheEarnedThisMonth,
		CurrentBalance:          totalMustacheUint,
		TotalOrders:             totalOrdersUint,
	})

}
