package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/carlosEA28/ai-code-review/internal/db"
	"github.com/carlosEA28/ai-code-review/internal/repository"
	"github.com/carlosEA28/ai-code-review/internal/service"

	"github.com/carlosEA28/ai-code-review/internal/web/server"
)

func main() {
	port := getenvOrDefault("PORT", "3000")
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	database, err := db.NewPostgresConnection(databaseURL)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer database.Close()

	userRepository := repository.NewUserRepository(database)
	webhookRepository := repository.NewWebhookRepository(database)
	userService := service.NewUserService(userRepository)
	webhookService := service.NewWebhookService(webhookRepository, os.Getenv("REVIEW_MODEL"))

	jwtService, err := service.NewJWTService(service.JWTConfig{
		SecretKey:  os.Getenv("JWT_SECRET"),
		Issuer:     os.Getenv("JWT_ISSUER"),
		Expiration: getenvDurationOrDefault("JWT_EXPIRATION", 24*time.Hour),
	})
	if err != nil {
		log.Fatal("Error creating jwt service: ", err)
	}

	authService, err := service.NewAuthService(userService, jwtService, service.GithubOAuthConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	})
	if err != nil {
		log.Fatal("Error creating auth service: ", err)
	}

	srv := server.NewServer(
		port,
		authService,
		jwtService,
		webhookService,
		os.Getenv("GITHUB_WEBHOOK_SECRET"),
	)

	if err = srv.Start(); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

func getenvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getenvDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return parsed
}
