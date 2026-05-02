.PHONY: build test lint vet frontend docker-build run-dev help validate validate-lattice-data-governai-routes validate-lattice-runtime-profile-routes validate-lattice-demo-command-bundle

## help: show this help message
help:
	@grep -E '^## [a-z]' Makefile | sed 's/^## /  /'

## build: compile the gateway binary
build:
	go build -o bin/gateway ./cmd/gateway

## test: run all Go tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## lint: run go vet (extend with golangci-lint if available)
lint: vet

## validate: validate fixture contracts
validate: validate-lattice-data-governai-routes validate-lattice-runtime-profile-routes validate-lattice-demo-command-bundle

validate-lattice-data-governai-routes:
	python3 tools/validate_lattice_data_governai_routes.py

validate-lattice-runtime-profile-routes:
	python3 tools/validate_lattice_runtime_profile_routes.py

validate-lattice-demo-command-bundle:
	python3 tools/validate_lattice_demo_command_bundle.py

## frontend: install npm deps and build the web UI bundle
frontend:
	cd web && npm ci && npm run build

## docker-build: build the multi-stage Docker image
docker-build:
	docker build -t cloudshell-fog:dev .

## run-dev: run the gateway in dev mode (stub connector, no OIDC)
run-dev: build
	USE_STUB_CONNECTOR=1 ./bin/gateway
