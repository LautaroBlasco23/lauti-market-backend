package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/database"
	apiInfrastructure "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

func main() {
	postgres, err := database.NewPostgres(database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "lauti_market"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgres.Close()

	db := postgres.DB()
	mux := http.NewServeMux()
	uuidGen := apiInfrastructure.NewUUIDGenerator()

	userModule := userinfra.Wire(mux, db, uuidGen)
	storeModule := storeinfra.Wire(mux, db, uuidGen)
	authinfra.Wire(mux, db, uuidGen, userModule, storeModule, authUtils.JwtConfig{
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration: 24 * time.Hour,
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
