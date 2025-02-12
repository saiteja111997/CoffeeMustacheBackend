package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	helper "coffeeMustacheBackend/pkg/helper"
	"coffeeMustacheBackend/pkg/server"
	"coffeeMustacheBackend/pkg/structures"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
)

var config structures.Config

func init() {

	if !helper.IsLambda() {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading environment variables file")
		}
	}

	config = structures.Config{
		DB_USERNAME:        os.Getenv("DB_USERNAME"),
		DB_PASSWORD:        os.Getenv("DB_PASSWORD"),
		DB_HOSTNAME:        os.Getenv("DB_HOSTNAME"),
		DB_PORT:            os.Getenv("DB_PORT"),
		DATABASE:           os.Getenv("DATABASE"),
		ORIGIN:             os.Getenv("ORIGIN"),
		TWILIO_ACCOUNT_SID: os.Getenv("TWILIO_ACCOUNT_SID"),
		TWILIO_AUTH_TOKEN:  os.Getenv("TWILIO_AUTH_TOKEN"),
		TWILIO_SERVICES_ID: os.Getenv("TWILIO_SERVICES_ID"),
		OPEN_AI_API_KEY:    os.Getenv("OPEN_AI_API_KEY"),
	}

	// Check if required variables are loaded
	if config.DB_HOSTNAME == "" {
		log.Fatalf("One or more required environment variables are missing")
	} else {
		fmt.Println("Successfully loaded environment variables from Lambda!")
	}
}

var fiberLambda *fiberadapter.FiberLambda

func main() {
	fmt.Println("Starting the server !!")
	app := fiber.New()

	// Use the CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.ORIGIN,
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	DB_USERNAME := config.DB_USERNAME
	DB_PASSWORD := config.DB_PASSWORD
	DB_HOSTNAME := config.DB_HOSTNAME
	DB_PORT := config.DB_PORT
	DATABASE := config.DATABASE

	//ctx := context.Background()
	if err := waitForHost(DB_HOSTNAME, DB_PORT); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Connection established")

	db, err := helper.Open(helper.Config{
		Username: DB_USERNAME,
		Password: DB_PASSWORD,
		Hostname: DB_HOSTNAME,
		Port:     DB_PORT,
		Database: DATABASE,
	})

	if err != nil {
		log.Println(err)
		return
	}

	db = db.Debug()
	db.AutoMigrate(&structures.User{}, &structures.Preference{}, &structures.MenuItem{}, &structures.ItemCustomization{}, &structures.CrossSell{}, &structures.Order{}, &structures.OrderItem{}, &structures.CuratedCart{}, &structures.CuratedCartItem{})
	fmt.Println("Auto migration done!!")

	defer db.Close()

	svr := server.Server{
		Config: config,
		Db:     db,
	}

	functionName := os.Getenv("FUNCTION_NAME")

	// Handle specific functions differently
	switch functionName {
	case "curatedCartCronJob":
		svr.RunCuratedCartsJob()
		return
	default:
		fmt.Println("Proceeding with normal server setup")
	}

	app.Get("/ping", svr.HealthCheck)
	app.Post("/sendOtp", svr.SendOtp)
	app.Post("/verifyOtp", svr.VerifyOtp)
	app.Post("/upsellItem", svr.UpsellItem)
	app.Post("/crossSellItem", svr.CrossSellItem)
	app.Post("/getUpsellAndCrossSell", svr.GetUpsellAndCrossSell)
	app.Post("/askMenuAI", svr.AskMenuAI)
	app.Post("/getMenu", svr.GetMenu)
	app.Post("/getFilteredList", svr.GetFilteredList)

	fmt.Println("Routing established!!")

	if helper.IsLambda() {
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
