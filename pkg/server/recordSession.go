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

	var sessionExists bool
	var role structures.UserRole

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
		} else {
			fmt.Println("Error fetching session:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to fetch session",
			})
		}
	} else {
		sessionExists = true
	}

	if sessionExists {
		role = structures.Guest
	} else {
		role = structures.Host
	}

	// Check if there is an existing user session for the given session ID
	var userSession structures.UserSession
	if err := s.Db.Where("session_id = ? AND user_id = ? AND status = ?", session.SessionID, userId, structures.UserActive).First(&userSession).Error; err != nil {
		// If no active user session exists, create a new user session

		// if record not found create a new session
		if err.Error() == "record not found" {
			newUserSession := structures.UserSession{
				UserSessionID: ksuid.New().String(),
				SessionID:     session.SessionID,
				UserID:        userId,
				Status:        structures.UserActive,
				JoinedAt:      time.Now(),
				Role:          role, // Default role as Guest or Host
			}

			if err := s.Db.Create(&newUserSession).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create user session",
				})
			}

			userSession = newUserSession
		} else {
			fmt.Println("Error fetching user session:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to fetch user session",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "User session recorded successfully",
		"session_id":      session.SessionID,
		"user_session_id": userSession.UserSessionID,
		"user_id":         userId,
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

	if err := s.Db.Where("session_id =?", req.SessionID).First(&session).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	sessionStatus := session.SessionStatus == structures.Active
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session_status": sessionStatus,
	})

}
