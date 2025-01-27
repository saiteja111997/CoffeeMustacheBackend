package server

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

func (s *Server) VerifyOtp(c *fiber.Ctx) error {

	phoneNumber := c.FormValue("phone_number")
	code := c.FormValue("code")

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
