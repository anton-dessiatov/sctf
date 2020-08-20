COMPOSE=docker-compose

.PHONY: db-down
db-down:
	@echo "==> DB Down"
	@$(COMPOSE) down --remove-orphans
	@docker kill sctf_db 2>/dev/null || true

.PHONY: db-up
db-up:
	@echo "==> DB Up"
	@$(COMPOSE) up --build -d db_check

.PHONY: db-migrate
db-migrate: db-up
	@echo "==> Migrate db changes"
	@goose -dir dal/schema/ mysql "root:101202@tcp(127.0.0.1:5556)/terra?parseTime=true" up

.PHONY: build
build: providers
	@echo "==> Building development binary..."
	@go version
	@go build

.PHONY: tools
tools:
	cd tools && \
		go install "github.com/pressly/goose/cmd/goose"

.PHONY: providers
providers: terraform-provider-aws terraform-provider-google

terraform-provider-aws:
	DIR=$$(pwd) && \
	cd providers/aws && \
	go install "github.com/terraform-providers/terraform-provider-aws" && \
	cp "$$GOPATH/bin/terraform-provider-aws" "$$DIR"

terraform-provider-google:
	DIR=$$(pwd) && \
	cd providers/google && \
	go install "github.com/hashicorp/terraform-provider-google" && \
	cp "$$GOPATH/bin/terraform-provider-google" "$$DIR"
