.PHONY: help setup dev build test run \
        docker-build docker-run docker-stop \
        compose-up compose-down compose-logs \
        bootstrap

IMAGE ?= evermeet
DATA  ?= $(PWD)/data

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development"
	@echo "  setup          Install web dependencies"
	@echo "  dev            Start Go + Vite dev servers"
	@echo "  build          Build web assets and Go binary"
	@echo "  test           Run Go tests"
	@echo "  run            Run local binary (requires build)"
	@echo "  bootstrap      Run local bootstrap-only node on port 4002"
	@echo ""
	@echo "Docker"
	@echo "  docker-build   Build Docker image (tag: $(IMAGE))"
	@echo "  docker-run     Run image on ports 7331/4001 with ./data volume"
	@echo "  docker-stop    Stop and remove the container"
	@echo ""
	@echo "Docker Compose"
	@echo "  compose-up     Start all services (detached)"
	@echo "  compose-down   Stop and remove all services"
	@echo "  compose-logs   Follow logs for all services"

# ---- development ----

setup:
	cd web && npm install

dev:
	@echo "Starting dev servers..."
	@(cd web && npm run dev) &
	@go run . --config evermeet.toml

build:
	cd web && npm run build
	go build -o evermeet .

test:
	go test ./...

run:
	./evermeet --config evermeet.toml

bootstrap:
	go run . --bootstrap --p2p-port 4002 --data ./data_bootstrap

# ---- docker ----

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm --name $(IMAGE) \
		-p 7331:7331 \
		-p 4001:4001 \
		-v $(DATA):/data \
		$(IMAGE)

docker-stop:
	docker stop $(IMAGE) 2>/dev/null || true
	docker rm $(IMAGE) 2>/dev/null || true

# ---- docker compose ----

compose-up:
	docker compose up -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f
