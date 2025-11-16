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

.PHONY: mocks
mocks:
	mockgen -source=internal/domain/member/interfaces/member-repo.go \
	-destination=internal/domain/member/mocks/mock-member-repo.go
	mockgen -source=internal/domain/team/interfaces/team-repo.go \
	-destination=internal/domain/team/mocks/mock-team-repo.go
	mockgen -source=internal/domain/pull-request/interfaces/pull-request-repo.go \
	-destination=internal/domain/pull-request/mocks/mock-pull-request-repo.go
	mockgen -source=internal/domain/statistics/interfaces/stats-repo.go \
	-destination=internal/domain/statistics/mocks/mock-stats-repo.go

.PHONY: test
test: 
	go test -v -count=1 ./...

.PHONY: test10
test10: 
	go test -v -count=10 ./...

cover:
	go test -coverprofile=cover.out -count=1 ./...
	go tool cover -html=cover.out
	DEL cover.out

.PHONY: load_test_data
load_test_data:
	docker exec -i pr-svc_postgres psql -U Admin -d pr-service -f /scripts/fill-with-test-data.sql