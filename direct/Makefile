COMPOSE=docker-compose

.PHONY: db-down
db-down:
	@echo "==> DB Down"
	@$(COMPOSE) down --remove-orphans
	@docker kill direct_db 2>/dev/null || true

.PHONY: db-up
db-up:
	@echo "==> DB Up"
	@$(COMPOSE) up --build -d db_check

.PHONY: db-migrate
db-migrate: db-up
	@echo "==> Migrate db changes"
	@goose -dir dal/schema/ mysql "root:101202@tcp(127.0.0.1:5558)/direct?parseTime=true" up

.PHONY: build
build:
	@echo "==> Building development binary..."
	@go version
	@go build

.PHONY: tools
tools:
	cd tools && \
		go install "github.com/pressly/goose/cmd/goose"

