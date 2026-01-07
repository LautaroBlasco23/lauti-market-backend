package main

import (
	"log"
	"net/http"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

func main() {
	mux := http.NewServeMux()
	uuidGen := api.NewUUIDGenerator()

	userModule := userinfra.Wire(mux, uuidGen)

	authinfra.Wire(mux, uuidGen, userModule, authinfra.Config{
		JWTSecret:     "your-secret-key-change-in-production",
		JWTExpiration: 24 * time.Hour,
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
