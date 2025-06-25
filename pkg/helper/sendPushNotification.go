package helper

import (
	"context"
	"encoding/base64"
	"log"
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
				Title: "Congratulations!!",
				Body:  "You have just implement push notification",
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

func getDecodedFireBaseKey() ([]byte, error) {

	fireBaseAuthKey := os.Getenv("FIREBASE_AUTH_KEY")

	decodedKey, err := base64.StdEncoding.DecodeString(fireBaseAuthKey)
	if err != nil {
		return nil, err
	}

	return decodedKey, nil
}
