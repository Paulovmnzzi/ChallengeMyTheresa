package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/app/health"
	"github.com/mytheresa/go-hiring-challenge/models"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, close := database.New(
		requireEnv("POSTGRES_USER"),
		requireEnv("POSTGRES_PASSWORD"),
		requireEnv("POSTGRES_DB"),
		requireEnv("POSTGRES_PORT"),
	)
	defer close()

	// Initialize handlers
	prodRepo := models.NewProductsRepository(db)
	cat := catalog.NewCatalogHandler(prodRepo)

	catRepo := models.NewCategoriesRepository(db)
	cats := categories.NewCategoriesHandler(catRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.HandleGet)
	mux.HandleFunc("GET /catalog", cat.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", cat.HandleGetByCode)
	mux.HandleFunc("GET /categories", cats.HandleGet)
	mux.HandleFunc("POST /categories", cats.HandlePost)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", requireEnv("HTTP_PORT")),
		Handler: mux,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
