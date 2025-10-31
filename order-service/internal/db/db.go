package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"order-service/internal/logger"
	"order-service/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
	log logger.Logger
}

// Connect подключается к Postgres с retry и логирует прогресс
func Connect(dsn string, log logger.Logger) (*DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				log.Info("Postgres is ready")
				return &DB{db, log}, nil
			} else {
				err = pingErr
			}
		}
		log.Error(fmt.Sprintf("Postgres not ready yet (%d/10)", i+1), err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to Postgres: %v", err)
}

// InsertOrder вставляет один заказ в базу
func (db *DB) InsertOrder(o models.Order) error {
	data, err := json.Marshal(o)
	if err != nil {
		db.log.Error("Failed to marshal order "+o.OrderUID, err)
		return err
	}

	_, err = db.Exec(
		`INSERT INTO orders(order_uid, data) VALUES($1, $2)
		 ON CONFLICT (order_uid) DO NOTHING`,
		o.OrderUID, data,
	)
	if err != nil {
		db.log.Error("Failed to insert order "+o.OrderUID, err)
	}
	return err
}

// GetOrder возвращает заказ по order_uid
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

// LoadAllOrders возвращает все заказы из базы
func (db *DB) LoadAllOrders() ([]models.Order, error) {
	rows, err := db.Query("SELECT data FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			db.log.Error("Scan error", err)
			continue
		}
		var o models.Order
		if err := json.Unmarshal(data, &o); err != nil {
			db.log.Error("Unmarshal error", err)
			continue
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// CreateTable создаёт таблицу orders, если её нет
func (db *DB) CreateTable() error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS orders(
            order_uid TEXT PRIMARY KEY,
            data JSONB
        )
    `)
	if err != nil {
		db.log.Error("Failed to create table", err)
	}
	return err
}

// InsertOrders вставляет список заказов
func (db *DB) InsertOrders(orders []models.Order) {
	for _, o := range orders {
		if err := db.InsertOrder(o); err != nil {
			// уже залогировано внутри InsertOrder
			continue
		}
	}
}
