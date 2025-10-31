package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"order-service/internal/models"

	_ "github.com/lib/pq"
)

type DB struct{ *sql.DB }

func Connect(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil { return nil, err }
	if err := db.Ping(); err != nil { return nil, err }
	return &DB{db}, nil
}

func (db *DB) InsertOrder(o models.Order) error {
	data, err := json.Marshal(o)
	if err != nil { return err }

	_, err = db.Exec(
		`INSERT INTO orders(order_uid, data) VALUES($1, $2)
		 ON CONFLICT (order_uid) DO NOTHING`,
		o.OrderUID, data,
	)
	return err
}

func (db *DB) GetOrder(orderUID string) (models.Order, error) {
	var data []byte
	if err := db.QueryRow("SELECT data FROM orders WHERE order_uid = $1", orderUID).Scan(&data); err != nil {
		return models.Order{}, err
	}
	var o models.Order
	if err := json.Unmarshal(data, &o); err != nil {
		return models.Order{}, err
	}
	return o, nil
}

func (db *DB) LoadAllOrders() ([]models.Order, error) {
	rows, err := db.Query("SELECT data FROM orders")
	if err != nil { return nil, err }
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		var o models.Order
		if err := json.Unmarshal(data, &o); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (db *DB) CreateTable() error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS orders(
            order_uid TEXT PRIMARY KEY,
            data JSONB
        )
    `)
    return err
}

// helper: вставка списка заказов
func (db *DB) InsertOrders(orders []models.Order) {
    for _, o := range orders {
        if err := db.InsertOrder(o); err != nil {
            log.Printf("failed to insert order %s: %v", o.OrderUID, err)
        }
    }
}

