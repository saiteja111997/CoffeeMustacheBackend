package server

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

func (s *Server) VerifyOtp(c *fiber.Ctx) error {

	var request struct {
		PhoneNumber string `json:"phone_number"`
		Code        string `json:"code"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
		})
	}

	phoneNumber := request.PhoneNumber
	code := request.Code

	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(phoneNumber)
	params.SetCode(code)

	var client *twilio.RestClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.Config.TWILIO_ACCOUNT_SID,
		Password: s.Config.TWILIO_AUTH_TOKEN,
	})

	resp, err := client.VerifyV2.CreateVerificationCheck(s.Config.TWILIO_SERVICES_ID, params)
	if err != nil {
		return err
	}

	// BREAKING CHANGE IN THE VERIFY API
	// https://www.twilio.com/docs/verify/quickstarts/verify-totp-change-in-api-response-when-authpayload-is-incorrect
	if *resp.Status != "approved" {
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": errors.New("not a valid code"),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
	})
}
