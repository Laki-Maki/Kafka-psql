package kafka

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/cache"
	"order-service/internal/db"
	"order-service/internal/models"

	"github.com/segmentio/kafka-go"
)

func StartConsumer(ctx context.Context, dbConn *db.DB, c *cache.Cache, brokers []string, topic, groupID string) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	go func() {
		defer r.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("Kafka consumer stopped")
				return
			default:
				msg, err := r.ReadMessage(ctx)
				if err != nil {
					log.Printf("kafka read error: %v", err)
					continue
				}

				var order models.Order
				if err := json.Unmarshal(msg.Value, &order); err != nil {
					log.Printf("invalid order JSON: %v", err)
					continue
				}

				// 🔹 защищаем Validate от panic
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered from panic in order.Validate: %v", r)
						}
					}()
					if err := order.Validate(); err != nil {
						log.Printf("order validation failed: %v", err)
						return
					}

					// сохраняем только если Validate прошла
					if err := dbConn.InsertOrder(order); err != nil {
						log.Printf("db insert error: %v", err)
						return
					}
					c.Set(order)
					log.Printf("order %s saved & cached", order.OrderUID)
				}()
			}
		}
	}()
	return r
}
