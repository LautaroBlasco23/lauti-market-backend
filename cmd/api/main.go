package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/database"
	apiInfrastructure "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	postgres, err := database.NewPostgres(database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "lauti_market"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
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

	server := &http.Server{
		Addr:         getEnv("PORT", ":8000"),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
