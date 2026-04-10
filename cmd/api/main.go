package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/database"
	apiInfrastructure "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	imageinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/image/infrastructure"
	orderinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure"
	paymentinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure"
	productinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/mercadopago"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

func main() {
	// Initialize pretty logger with colored output
	apiInfrastructure.InitLogger()

	if err := run(); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	postgres, err := database.NewPostgres(&database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "lauti_market"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		if closeErr := postgres.Close(); closeErr != nil {
			slog.Info("error closing database connection", slog.String("error", closeErr.Error()))
		}
	}()

	db := postgres.DB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	uuidGen := apiInfrastructure.NewUUIDGenerator()

	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")

	jwtGen := authUtils.NewJWTGenerator(jwtSecret, 24*time.Hour)
	authMw := apiInfrastructure.NewAuthMiddleware(jwtGen)

	userModule := userinfra.Wire(mux, db, uuidGen, authMw)

	mpOAuth := mercadopago.NewOAuthClient(
		getEnv("MP_CLIENT_ID", ""),
		getEnv("MP_CLIENT_SECRET", ""),
		getEnv("MP_REDIRECT_URI", ""),
	)
	storeModule := storeinfra.Wire(mux, db, uuidGen, authMw, mpOAuth)

	authinfra.Wire(mux, db, uuidGen, userModule, storeModule, authUtils.JwtConfig{
		JWTSecret:     jwtSecret,
		JWTExpiration: 24 * time.Hour,
	}, authMw)

	imageModule, imgErr := imageinfra.Wire(getEnv("IMAGE_STORE_ADDR", "localhost:50051"))
	if imgErr != nil {
		return fmt.Errorf("connecting to image service: %w", imgErr)
	}
	defer func() {
		if closeErr := imageModule.Close(); closeErr != nil {
			slog.Info("error closing image module", slog.String("error", closeErr.Error()))
		}
	}()

	productModule := productinfra.Wire(mux, db, uuidGen, storeModule.Repository, imageModule.Client, authMw)
	orderModule := orderinfra.Wire(mux, db, uuidGen, productModule.Repository, authMw)

	paymentinfra.Wire(mux, db, uuidGen, orderModule.Repository, productModule.Repository, storeModule.Service, authMw,
		getEnv("MERCADO_PAGO_ACCESS_TOKEN", ""),
		getEnv("MERCADO_PAGO_WEBHOOK_SECRET", ""),
		getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
		getEnv("MERCADO_PAGO_NOTIFICATION_URL", ""),
	)

	// Chain middleware: Logging -> CORS -> Router
	// Logging is outermost to capture all requests including CORS preflight
	handler := apiInfrastructure.LoggingMiddleware(apiInfrastructure.CORSMiddleware(mux))

	server := &http.Server{
		Addr:         getEnv("PORT", ":8000"),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("shutting down server")
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

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
