// Package main is the entry point for the order-service.
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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Release-Orchestrator/order-service/internal/client"
	"github.com/Release-Orchestrator/order-service/internal/config"
	"github.com/Release-Orchestrator/order-service/internal/handler"
	"github.com/Release-Orchestrator/order-service/internal/repository"
	"github.com/Release-Orchestrator/order-service/internal/service"
	"github.com/Release-Orchestrator/order-service/migrations"
)

func main() {
	cfg := config.Load()

	dbpool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}
	log.Println("database connected")

	if err := runMigrations(dbpool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")

	repo := repository.NewOrderRepository(dbpool)
	userClient := client.NewUserServiceClient(cfg.UserServiceURL)
	paymentClient := client.NewPaymentServiceClient(cfg.PaymentServiceURL)
	svc := service.NewOrderService(repo, userClient, paymentClient)
	orderHandler := handler.NewOrderHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	api := r.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exited")
}

func runMigrations(db *pgxpool.Pool) error {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if len(entry.Name()) < 5 || entry.Name()[len(entry.Name())-4:] != ".sql" {
			continue
		}
		data, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		if _, err := db.Exec(context.Background(), string(data)); err != nil {
			return err
		}
		log.Printf("applied migration: %s", entry.Name())
	}

	return nil
}
