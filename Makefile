.PHONY: restart
restart: clean up

.PHONY: rebuild
rebuild: build refresh

.PHONY: refresh
refresh: down up

.PHONY: build
build:
	docker compose build

.PHONY: up
up: 
	docker compose up -d

.PHONY: down
down: 
	docker compose down 

.PHONY: clean
clean:
	docker compose down -v --remove-orphans

.PHONY: swag
swag:
	swag init -g internal/presentation/rest/gin/routes.go --output docs --parseDependency true