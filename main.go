package main

import (
	"context"
	"fmt"
	"net"
	"time"

	helpers "coffeeMustacheBackend/pkg/helper"
	"coffeeMustacheBackend/pkg/server"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
)

var fiberLambda *fiberadapter.FiberLambda

func main() {
	fmt.Println("Starting the server !!")
	app := fiber.New()

	svr := server.Server{}

	app.Get("/ping", svr.HealthCheck)

	fmt.Println("Routing established!!")

	if helpers.IsLambda() {
		fiberLambda = fiberadapter.New(app)
		lambda.Start(Handler)
	} else {
		fmt.Println("Starting server locally!!")
		err := app.Listen(":8080")

		if err != nil {
			fmt.Println("An error occured while starting the server : ", err)
		}
	}
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Proxy the request to the Fiber app and get the response
	response, err := fiberLambda.ProxyWithContext(ctx, request)

	response.Headers = make(map[string]string)

	// Add CORS headers to the response
	response.Headers["Access-Control-Allow-Origin"] = "*"
	response.Headers["Access-Control-Allow-Methods"] = "GET,POST,PUT,DELETE"
	response.Headers["Access-Control-Allow-Headers"] = "Origin, Content-Type, Accept"
	response.Headers["Access-Control-Allow-Credentials"] = "true"

	return response, err
}

func waitForHost(host, port string) error {
	timeOut := time.Second

	if host == "" {
		return errors.Errorf("unable to connect to %v:%v", host, port)
	}

	for i := 0; i < 60; i++ {
		fmt.Printf("waiting for %v:%v ...\n", host, port)
		conn, err := net.DialTimeout("tcp", host+":"+port, timeOut)
		if err == nil {
			fmt.Println("done!")
			conn.Close()
			return nil
		}

		time.Sleep(time.Second)
	}

	return errors.Errorf("timeout attempting to connect to %v:%v", host, port)
}
