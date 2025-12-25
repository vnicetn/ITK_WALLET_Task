package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itk/wallet/internal/handlers"
	"github.com/itk/wallet/internal/pkg/postgres"
	"github.com/itk/wallet/internal/repository"
	"github.com/itk/wallet/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Failed to load config.env: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := postgres.NewPostgresDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	walletRepo := repository.NewWalletRepository(db)
	walletService := service.NewWalletService(walletRepo)
	walletHandler := handlers.NewWalletHandler(walletService)

	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/wallet", walletHandler.ProcessOperation)
		v1.GET("/wallets/:WALLET_UUID", walletHandler.GetBalance)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shuts down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
