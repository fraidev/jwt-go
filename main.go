package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	jwt "github.com/form3tech-oss/jwt-go"
	jwtware "github.com/gofiber/jwt/v2"
)

var (
	// Obviously, this is just a test example. Do not do this in production.
	// In production, you would have the private key and public key pair generated
	// in advance. NEVER add a private key to any GitHub repo.
	privateKey *rsa.PrivateKey
)

func main() {
	app := fiber.New()

    dir, err := os.Getwd()
    if err != nil {
            log.Fatal(err)
    }
    fmt.Println(dir)
	
	privateKey, err = rsaConfigSetup(dir + "/private.pem", dir + "/public.pem")
	if err != nil {
		log.Fatalf("rsa.GenerateKey: %v", err)
	}

	// Login route
	app.Post("/login", login)

	// Unauthenticated route
	app.Get("/", accessible)

	pub := privateKey.Public();

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningMethod: "RS256",
		SigningKey:    pub,
	}))

	// Restricted Routes
	app.Get("/restricted", restricted)

	app.Listen(":3000")
}

func login(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("pass")

	// Throws Unauthorized error
	if user != "john" || pass != "doe" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodRS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("token.SignedString: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func accessible(c *fiber.Ctx) error {
	return c.SendString("Accessible")
}

func restricted(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}