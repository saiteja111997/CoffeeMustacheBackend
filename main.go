package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	helper "coffeeMustacheBackend/pkg/helper"
	"coffeeMustacheBackend/pkg/server"
	"coffeeMustacheBackend/pkg/structures"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt"
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
		JWT_SECRET:         os.Getenv("JWT_SECRET"),
	}

	// Check if required variables are loaded
	if config.DB_HOSTNAME == "" {
		log.Fatalf("One or more required environment variables are missing")
	} else {
		fmt.Println("Successfully loaded environment variables from Lambda!")
	}
}

var fiberLambda *fiberadapter.FiberLambda

// Middleware to extract and validate JWT
func ExtractJWT(c *fiber.Ctx) error {
	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing Authorization header",
		})
	}

	// Check if it's a Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Authorization header format",
		})
	}

	// Extract token by removing "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.JWT_SECRET), nil
	})

	// Check for parsing or validation errors
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid or expired token",
		})
	}

	// Extract claims (payload) from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to extract token claims",
		})
	}

	// Example: Extract user phone number from claims
	userId := claims["user_id"].(float64)
	fmt.Println("User ID extracted from token: ", userId)

	// Store phone number in context for further use
	c.Locals("userId", userId)

	// Proceed to next handler
	return c.Next()
}

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
	db.AutoMigrate(&structures.User{}, &structures.Preference{}, &structures.MenuItem{}, &structures.ItemCustomization{}, &structures.CrossSell{}, &structures.CuratedCart{}, &structures.CuratedCartItem{}, &structures.Session{}, &structures.UserSession{}, &structures.Cart{}, &structures.CartItem{}, &structures.Order{}, &structures.OrderItem{}, &structures.Order{}, &structures.OrderItem{}, &structures.UpdateCartResult{})
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
	// Apply JWT middleware to protected routes
	app.Post("/upsellItem", ExtractJWT, svr.UpsellItem)
	app.Post("/getUpsellAndCrossSell", ExtractJWT, svr.GetUpsellAndCrossSell)
	app.Post("/askMenuAI", ExtractJWT, svr.AskMenuAI)
	app.Post("/getMenu", ExtractJWT, svr.GetMenu)
	app.Post("/getFilteredList", ExtractJWT, svr.GetFilteredList)
	app.Post("/getCrossSellData", ExtractJWT, svr.GetCrossSellData)
	app.Post("/recordUserSession", ExtractJWT, svr.RecordUserSession)
	app.Post("/getCuratedCart", ExtractJWT, svr.GetCuratedCart)
	app.Post("/addToCart", ExtractJWT, svr.AddToCart)
	app.Post("/getCart", ExtractJWT, svr.GetCart)
	app.Post("/updateCustomizations", ExtractJWT, svr.UpdateCustomizations)
	app.Post("/updateCrossSellItems", ExtractJWT, svr.UpdateCrossSellItems)
	app.Post("/updateQuantity", ExtractJWT, svr.UpdateQuantity)
	app.Post("/crossSellCheckout", ExtractJWT, svr.GetCheckoutCrossSells)
	app.Post("/upgradeCart", ExtractJWT, svr.UpgradeCart)
	app.Post("getItemAudio", ExtractJWT, svr.GetItemAudio)

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
