package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	"order-service/internal/models"

	
)

func main() {
	cfg := config.LoadConfig("config/config.yaml") // путь исправлен

	dbConn, err := db.Connect(fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.SSLMode,
	))
	if err != nil {
		log.Fatal(err)
	}

	// создаём таблицу, если нет
	if err := dbConn.CreateTable(); err != nil {
		log.Fatal(err)
	}

	c := cache.New(cfg.Cache.Size)
	orders, _ := dbConn.LoadAllOrders()
	c.Load(orders)
	log.Printf("cache warmup: %d orders loaded", c.Len())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kafkaReader := kafka.StartConsumer(ctx, dbConn, c, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)

	mux := http.NewServeMux()
	mux.HandleFunc("/order/", orderHandler(c))
	mux.Handle("/", http.FileServer(http.Dir("./internal/web")))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: mux}

	go func() {
		log.Printf("HTTP server listening on :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}
	cancel()
	kafkaReader.Close()
	dbConn.Close()
	log.Println("server exited gracefully")

	// загружаем sample JSON
	sampleData, err := os.ReadFile("assets/sample_order.json")
	if err != nil {
		log.Printf("failed to read sample orders: %v", err)
	} else {
		var sampleOrders []models.Order
		if err := json.Unmarshal(sampleData, &sampleOrders); err != nil {
			log.Printf("failed to unmarshal sample orders: %v", err)
		} else {
			dbConn.InsertOrders(sampleOrders)
			c.Load(sampleOrders)
			log.Printf("loaded %d sample orders into cache & db", len(sampleOrders))
		}
	}

}

func orderHandler(c *cache.Cache) http.HandlerFunc {
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
