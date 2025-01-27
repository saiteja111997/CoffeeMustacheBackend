package server

import (
	"coffeeMustacheBackend/pkg/structures"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

func (s *Server) SendOtp(c *fiber.Ctx) error {

	phoneNumber := c.FormValue("phone_number")
	name := c.FormValue("name")
	gender := c.FormValue("gender")

	// Validate input
	if phoneNumber == "" || name == "" || gender == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "All fields (phone_number, name, gender) are required",
		})
	}

	//Insert into database
	user := &structures.User{Phone: phoneNumber, Name: name, Gender: gender}
	err := s.Db.Create(user).Error
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err,
		})
	}

	params := &twilioApi.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")

	var client *twilio.RestClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.Config.TWILIO_ACCOUNT_SID,
		Password: s.Config.TWILIO_AUTH_TOKEN,
	})

	fmt.Println("Printing Services SID : ", s.Config.TWILIO_SERVICES_ID)
	resp, err := client.VerifyV2.CreateVerification(s.Config.TWILIO_SERVICES_ID, params)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"sid":   resp.Sid,
		"error": nil,
	})
}
