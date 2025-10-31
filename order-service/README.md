 Быстрый старт

Клонируем репозиторий

git clone <repo-url>
cd <repo-folder>


Поднимаем сервисы через Docker Compose

docker-compose up -d


Сервисы, которые поднимаются:

order-service-app — основной Go-сервис

order-service-kafka-1 — Kafka брокер

order-service-zookeeper-1 — Zookeeper

order-service-postgres-1 — PostgreSQL

Создание топиков в Kafka (если не создались автоматически)

docker exec -it order-service-kafka-1 /bin/sh
/opt/kafka/bin/kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092


Отправка тестового сообщения в топик

$json | ForEach-Object {
    $message = $_ | ConvertTo-Json -Compress
    $message | docker exec -i order-service-kafka-1 /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic orders
}


Проверка, что сообщения приходят

docker exec -it order-service-kafka-1 /bin/sh
/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic orders --from-beginning

 Структура проекта
├── models/       # Структуры данных (Order, Payment, Delivery, Item)
├── assets/       # JSON-файлы с тестовыми заказами
├── main.go       # Основной файл сервиса
├── docker-compose.yml
└── README.md

 Валидация данных

Для полей slice и map больше не используется dive для Delivery и Payment, чтобы избежать ошибок can't dive on a non slice or map.

 Логи

Kafka генерирует большое количество логов при старте из-за инициализации всех координаторов групп и загрузки __consumer_offsets. Это нормально и не является ошибкой.
