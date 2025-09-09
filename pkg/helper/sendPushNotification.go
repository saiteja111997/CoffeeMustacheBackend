package helper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

func SendPushNotification(deviceTokens []string, title, body string) error {
	decodedKey, err := getDecodedFireBaseKey()

	if err != nil {
		return err
	}

	opts := []option.ClientOption{option.WithCredentialsJSON(decodedKey)}

	app, err := firebase.NewApp(context.Background(), nil, opts...)

	if err != nil {
		log.Printf("Error in initializing firebase : %s", err)
		return err
	}

	fcmClient, err := app.Messaging(context.Background())

	if err != nil {
		log.Printf("Error in initializing firebase messaging client : %s", err)
		return err
	}

	for _, token := range deviceTokens {
		response, err := fcmClient.Send(context.Background(), &messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Token: token,
		})
		if err != nil {
			log.Printf("Failed to send message to token %s: %v", token, err)
			continue
		}
		log.Printf("Successfully sent message to token %s: %s", token, response)
	}
	return nil
}

type ExpoPushMessage struct {
	To               string                 `json:"to"`
	Vibrate          string                 `json:"vibrate"`
	VibrationPattern []int                  `json:"vibrationPattern"`
	Sound            string                 `json:"sound"`
	Title            string                 `json:"title"`
	Body             string                 `json:"body"`
	Priority         string                 `json:"priority"`
	Data             map[string]interface{} `json:"data"`
	ChannelID        string                 `json:"channelId"`
}

func SendExpoPushNotification(deviceTokens []string, title, body, url string) error {

	fmt.Println("Sending Expo Push Notification to tokens:", deviceTokens)

	for _, token := range deviceTokens {
		msg := ExpoPushMessage{
			To:               token,
			Vibrate:          "true",
			VibrationPattern: []int{0, 250, 250, 250},
			Sound:            "notification_sound.wav",
			Title:            title,
			Body:             body,
			Priority:         "high",
			Data: map[string]interface{}{
				"url":   "https://admin.coffeemustache.in/alerts/waiter-view?tab=new-orders",
				"extra": "data",
			},
			ChannelID: "custom_channel",
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send", bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to send push notification, status: %s", resp.Status)
		}
	}
	return nil
}

func getDecodedFireBaseKey() ([]byte, error) {

	fireBaseAuthKey := os.Getenv("FIREBASE_AUTH_KEY")

	decodedKey, err := base64.StdEncoding.DecodeString(fireBaseAuthKey)
	if err != nil {
		return nil, err
	}

	return decodedKey, nil
}
