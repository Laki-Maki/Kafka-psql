# Order Service

Сервис для обработки заказов с использованием Kafka и PostgreSQL.

## Требования

- Docker & Docker Compose
- Go 1.21+
- Make (опционально)

## Быстрый старт

### 1. Клонируем репозиторий
```bash
git clone <repo-url>
cd <repo-folder>
```

2. Поднимаем сервисы через Docker Compose
```bash
Копировать код
docker-compose up -d
```
Сервисы, которые поднимаются:
```bash
order-service-app — основной Go-сервис

order-service-kafka-1 — Kafka брокер

order-service-zookeeper-1 — Zookeeper

order-service-postgres-1 — PostgreSQL
```
3. Создание топиков в Kafka (если не создались автоматически)
bash
```bash
docker exec -it order-service-kafka-1 /bin/sh
/opt/kafka/bin/kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092
```
4. Отправка тестового сообщения в топик
```bash
powershell
$json | ForEach-Object {
    $message = $_ | ConvertTo-Json -Compress
    $message | docker exec -i order-service-kafka-1 /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic orders
}
```
5. Проверка, что сообщения приходят
```bash
docker exec -it order-service-kafka-1 /bin/sh
/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic orders --from-beginning
```
Структура проекта
```bash
Копировать код
├── models/       # Структуры данных (Order, Payment, Delivery, Item)
├── assets/       # JSON-файлы с тестовыми заказами
├── main.go       # Основной файл сервиса
├── docker-compose.yml
└── README.md
```

Валидация данных
```bash
Для полей slice и map больше не используется dive для Delivery и Payment, чтобы избежать ошибок:
vbnet
can't dive on a non slice or map
```
Логи
```bash
Kafka генерирует большое количество логов при старте из-за инициализации координаторов групп и загрузки __consumer_offsets
```
