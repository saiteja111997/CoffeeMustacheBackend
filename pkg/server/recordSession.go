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
		TableId     string `json:"table_id"`
		CafeId      uint   `json:"cafe_id"`
		CompletePos bool   `json:"complete_pos"`
	}

	var req RecordSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Check if the given table ID exists for given cafe ID
	// var table structures.Table
	// if err := s.Db.Where("name =? AND cafe_id =?", req.TableId, req.CafeId).First(&table).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Table not found",
	// 	})
	// }

	status := true

	// Check if the cafe has a complete_pos true or false
	// var cafe structures.Cafe
	// if err := s.Db.Where("id = ?", req.CafeId).First(&cafe).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Cafe not found",
	// 	})
	// }

	if !req.CompletePos {
		// Get session details from the session table
		var session structures.Session
		if err := s.Db.Where("table_name = ? AND cafe_id = ? AND session_status = ?", req.TableId, req.CafeId, structures.Active).First(&session).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Session not found",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":    "User session recorded successfully",
			"session_id": session.SessionID,
			"user_id":    userId,
			"status":     status,
		})
	}

	// Check if there is an active session for the given table
	var session structures.Session
	if err := s.Db.Where("table_name = ? AND session_status = ? AND cafe_id = ?", req.TableId, "Active", req.CafeId).First(&session).Error; err != nil {
		// If no active session, create a new session with a unique session ID using ksuid

		// Generate a random 4-digit numeric code for the table
		tableCode := fmt.Sprintf("%04d", time.Now().UnixNano()%10000)

		// if record not found create a new session
		if err.Error() == "record not found" {
			newSession := structures.Session{
				SessionID:     ksuid.New().String(),
				TableName:     req.TableId,
				TableCode:     tableCode,
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

	if session.TableCode == "" {
		// Generate a random 4-digit numeric code for the table
		tableCode := fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
		session.TableCode = tableCode

		// Update the session with the new table code
		if err := s.Db.Model(&structures.Session{}).Where("session_id = ?", session.SessionID).Update("table_code", tableCode).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update session with table code",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "User session recorded successfully",
		"session_id": session.SessionID,
		"table_code": session.TableCode,
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

func (s *Server) VerifyTableCode(c *fiber.Ctx) error {
	type VerifyTableCodeRequest struct {
		TableCode string `json:"table_code"`
		SessionID string `json:"session_id"`
	}

	var req VerifyTableCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	var session structures.Session
	if err := s.Db.Where("session_id = ? AND table_code = ? AND session_status = ?", req.SessionID, req.TableCode, structures.Active).First(&session).Error; err != nil {
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Invalid table code or session ID",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify table code",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Table code verified successfully",
		"table":   session.TableCode,
	})
}
