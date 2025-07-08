KAFKAFILLER = ./cmd/eventhandler/testhelpers/kafkafiller/main.go
EVENTHANDLER = ./cmd/eventhandler/main.go
HTTPSERVER = ./cmd/httpserver/main.go
WEBSITE = ./cmd/website/main.go


DOCKER_COMPOSE = docker-compose.yml


include .env
export


all: up-zookeeper up-kafka up-postgres run-kafkafiller run-eventhandler run-httpserver run-website

up-zookeeper:
	@echo "🚀 Запуск Zookeeper..."
	docker compose -f $(DOCKER_COMPOSE) up -d zookeeper
	@echo "⏳ Ожидание Zookeeper..."
	@until docker compose logs zookeeper 2>&1 | grep -q "binding to port"; do sleep 1; done
	@echo "✅ Zookeeper готов."


up-kafka: up-zookeeper
	@echo "🚀 Запуск Kafka..."
	docker compose -f $(DOCKER_COMPOSE) up -d kafka
	@echo "⏳ Ожидание Kafka..."
#	@until docker compose logs kafka 2>&1 | grep -q "started \(kafka.server.KafkaServer\)"; do sleep 1; done
	@until nc -z localhost 9092; do sleep 1; done
	@echo "✅ Kafka готов."

up-postgres:
	@echo "🚀 Запуск PostgreSQL..."
	docker compose -f $(DOCKER_COMPOSE) up -d postgres
	@echo "⏳ Ожидание PostgreSQL..."
	@until docker compose logs postgres | grep "database system is ready to accept connections"; do sleep 1; done
	@echo "✅ PostgreSQL готов."

run-kafkafiller:
	@echo "▶️ Запуск kafkafiller..."
	go run $(KAFKAFILLER)

run-eventhandler:
	@echo "▶️ Запуск eventhandler..."
	nohup go run $(EVENTHANDLER) > logs/eventhandler.log 2>&1 &

run-httpserver:
	@echo "▶️ Запуск httpserver..."
	nohup go run $(HTTPSERVER) > logs/httpserver.log 2>&1 &

run-website:
	@echo "▶️ Запуск website..."
	go run $(WEBSITE)

down:
	docker compose -f $(DOCKER_COMPOSE) down

