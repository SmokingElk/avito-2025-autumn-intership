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

.PHONY: test
test: 
	go test -v -count=1 ./...

cover:
	go test -coverprofile=cover.out -count=1 ./...
	go tool cover -html=cover.out
	DEL cover.out