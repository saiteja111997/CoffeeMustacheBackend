package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
)

func (s *Server) RecordUserSession(c *fiber.Ctx) error {
	// Get user ID from locals
	userId := uint(c.Locals("userId").(float64))

	// Parse table ID from request body
	type RecordSessionRequest struct {
		TableId string `json:"table_id"`
		CafeId  uint   `json:"cafe_id"`
	}

	var req RecordSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if the given table ID exists for given cafe ID
	var table structures.Table
	if err := s.Db.Where("name =? AND cafe_id =?", req.TableId, req.CafeId).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	status := true

	// Check if the cafe has a complete_pos true or false
	// var cafe structures.Cafe
	// if err := s.Db.Where("id = ?", req.CafeId).First(&cafe).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Cafe not found",
	// 	})
	// }

	// Check if there is an active session for the given table
	var session structures.Session
	if err := s.Db.Where("table_name = ? AND session_status = ?", req.TableId, "Active").First(&session).Error; err != nil {
		// If no active session, create a new session with a unique session ID using ksuid

		// if record not found create a new session
		if err.Error() == "record not found" {
			newSession := structures.Session{
				SessionID:     ksuid.New().String(),
				TableName:     req.TableId,
				CafeID:        req.CafeId,
				SessionStatus: structures.Active,
				StartTime:     time.Now(),
				CreatedBy:     userId,
			}

			if err := s.Db.Create(&newSession).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create new session",
				})
			}

			// Assign the newly created session for further user session checks
			session = newSession
			status = false
		} else {
			fmt.Println("Error fetching session:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to fetch session",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "User session recorded successfully",
		"session_id": session.SessionID,
		"user_id":    userId,
		"status":     status,
	})
}

func (s *Server) InvalidateSession(c *fiber.Ctx) error {
	// Parse session ID from request body
	type InvalidateSessionRequest struct {
		SessionID string `json:"session_id"`
	}

	var req InvalidateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Find and update the session status to Inactive and set the end time
	if err := s.Db.Model(&structures.Session{}).
		Where("session_id = ?", req.SessionID).
		Updates(map[string]interface{}{
			"session_status": structures.Inactive,
			"end_time":       time.Now(),
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to invalidate session",
		})
	}

	// Invalidate the user sessions related to this session ID by setting LeftAt and updating the status
	if err := s.Db.Model(&structures.UserSession{}).
		Where("session_id = ? AND left_at IS NULL", req.SessionID).
		Updates(map[string]interface{}{
			"left_at": time.Now(),
			"status":  structures.UserInactive,
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to invalidate user sessions",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Session invalidated successfully",
	})
}

func (s *Server) CheckSessionStatus(c *fiber.Ctx) error {

	type CheckSessionStatusRequest struct {
		SessionID string `json:"session_id"`
	}

	var req CheckSessionStatusRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	var session structures.Session

	if result := s.Db.Where("session_id = ? AND session_status = ?", req.SessionID, "Active").First(&session); result.Error != nil {
		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"session_status": false,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session_status": true,
		"session_id":     session.SessionID,
	})
}

func (s *Server) GetSessionDetails(c *fiber.Ctx) error {
	cafeId := c.Locals("cafeId").(float64)
	tableName := c.Locals("tableName").(string)

	var session structures.Session
	if err := s.Db.Where("cafe_id = ? AND table_name = ?", cafeId, tableName).First(&session).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session_id":     session.SessionID,
		"table_name":     session.TableName,
		"cafe_id":        session.CafeID,
		"session_status": session.SessionStatus,
	})
}
