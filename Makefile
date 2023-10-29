export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: compose-up

compose-up-integration-test: ### Run docker-compose with integration test
	docker-compose up --build --abort-on-container-exit --exit-code-from integration
.PHONY: compose-up-integration-test

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

run: ### run
	go mod tidy && go mod download && \
	CONFIG_PATH="./config/config.yml" CGO_ENABLED=0 go run ./cmd/app --filepath=./integration-test/list.csv
.PHONY: run

docker-rm-volume: ### remove docker volume
	docker volume rm go-clean-template_pg-data
.PHONY: docker-rm-volume

lint: ### check by golangci linter
	golangci-lint run ./... --config=./.golangci.yml
.PHONY: lint

lint-fast: ### check by hadolint linter
	golangci-lint run ./... --fast --config=./.golangci.yml
.PHONY: lint-fast

test: ### run test
	go test -v -cover -race ./internal/...
.PHONY: test

integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

mock: ### run mockgen
	go generate ./...
.PHONY: mock

build: ### build for windows, linux & darwin GOOS, all is x64  
	env GOOS="linux" GOARCH="amd64" CGO_ENABLED=0 go build -o build/ctrl_linux -ldflags "-w -s" cmd/app/main.go
	env GOOS="darwin" GOARCH="amd64" CGO_ENABLED=0 go build -o build/ctrl_darwin -ldflags "-w -s" cmd/app/main.go
	env GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 go build -o build/ctrl_win64 -ldflags "-w -s" cmd/app/main.go
.PHONY: build

bin-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest