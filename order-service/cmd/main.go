package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"order-service/config"
	"order-service/internal/cache"
	"order-service/internal/db"
	"order-service/internal/kafka"
	"order-service/internal/logger"
	"order-service/internal/models"
)

func main() {
	log := &logger.StdLogger{}

	cfg := config.LoadConfig("config/config.yaml")

	// подключаем Postgres с retry
	dbConn, err := db.Connect(fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.SSLMode,
	), log)
	if err != nil {
		log.Error("Failed to connect Postgres", err)
		os.Exit(1)
	}

	if err := dbConn.CreateTable(); err != nil {
		log.Error("Failed to create table", err)
		os.Exit(1)
	}

	c := cache.New(cfg.Cache.Size)
	orders, _ := dbConn.LoadAllOrders()
	c.Load(orders)
	log.Info(fmt.Sprintf("Cache warmup: %d orders loaded", c.Len()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaReader := kafka.StartConsumer(ctx, dbConn, c, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	if kafkaReader == nil {
		log.Error("Kafka consumer did not start", fmt.Errorf("consumer nil"))
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/order/", orderHandler(c, log))
	mux.Handle("/", http.FileServer(http.Dir("./internal/web")))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: mux}

	go func() {
		log.Info(fmt.Sprintf("HTTP server listening on :%d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server listen error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown signal received")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error("Server forced to shutdown", err)
	}

	cancel()
	kafkaReader.Close()
	dbConn.Close()
	log.Info("Server exited gracefully")

	// загружаем sample JSON
	sampleData, err := os.ReadFile("assets/sample_order.json")
	if err != nil {
		log.Error("Failed to read sample orders", err)
	} else {
		var sampleOrders []models.Order
		if err := json.Unmarshal(sampleData, &sampleOrders); err != nil {
			log.Error("Failed to unmarshal sample orders", err)
		} else {
			dbConn.InsertOrders(sampleOrders)
			c.Load(sampleOrders)
			log.Info(fmt.Sprintf("Loaded %d sample orders into cache & db", len(sampleOrders)))
		}
	}
}

func orderHandler(c *cache.Cache, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/order/")
		if id == "" {
			http.Error(w, "order id required", http.StatusBadRequest)
			return
		}
		o, ok := c.Get(id)
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
			return
		}
		json.NewEncoder(w).Encode(o)
	}
}
