# Как запустить Шаг 1


1) Установите Go 1.22+
2) В корне выполните:


go mod tidy
go run ./cmd


3) Откройте браузер: http://localhost:8081
• Кнопка «Пример» сразу подставит order_uid из sample_order.json
• Или запрос через curl: curl http://localhost:8081/order/b563feb7b2b84b6test


# Что дальше (Шаг 2 и 3 — короткий анонс)
# • Шаг 2: добавим Postgres
# - internal/db/db.go (подключение)
# - internal/db/repository.go (Save/Get/All c JSONB)
# - при старте: загрузим все заказы из БД в кэш
# - в /order/<id> при промахе кэша — дотянем из БД
# • Шаг 3: добавим Kafka consumer
# - internal/kafka/consumer.go (segmentio/kafka-go)
# - чтение сообщений, валидация, транзакция: сохранить в БД -> положить в кэш -> commit offset
# - обработка ошибок, логирование некорректных сообщений