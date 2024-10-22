package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func HandleCors(ips []string) gin.HandlerFunc {
	// Define trusted origins

	trustedOrigins := []string{
		"https://127.0.0.1",
		"http://192.168.198.65:4200",
		"https://localhost:4200",
		"http://localhost:4200",
		"https://192.168.199.163:444",
		"https://192.168.0.107",
		"https://192.168.199.163",
		"https://192.168.199.163:443",
		"https://192.168.199.163:8082",
	}

	for _, ip := range ips {
		trustedOrigins = append(trustedOrigins, "https://"+ip)
	}
	fmt.Println("Trusted origins: ", trustedOrigins)

	// Configure CORS
	return cors.New(cors.Config{
		// AllowAllOrigins:           true,
		AllowOrigins:     trustedOrigins,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "UPDATE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Access-Control-Allow-Origin", "Content-Type", "Authorization", "x-requested-with", "x-forwarded-for"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "x-requested-with", "x-forwarded-for"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

}
