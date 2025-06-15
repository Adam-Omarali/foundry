// main.go
package main

import (
	"fmt"
	"foundry/backend/api"
	"foundry/backend/middleware"
	"foundry/backend/utils"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(".env.local"); err != nil {
		log.Fatal("Error loading .env file")
	} else {
		fmt.Println("Env file loaded")
	}

	// Initialize JWT
	if err := utils.InitializeJWT(); err != nil {
		log.Fatal("Error initializing JWT:", err)
	}

	// youtube.GetVideoData()

	http.HandleFunc("/api/signin", middleware.EnableCORS(api.SignInHandler))
	http.HandleFunc("/api/remember", middleware.EnableCORS(api.RememberHandler))
	http.HandleFunc("/api/verifytoken", middleware.EnableCORS(api.VerifyJWTHandler))
	http.HandleFunc("/api/documents", middleware.EnableCORS(api.GetDocumentsHandler))
	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
