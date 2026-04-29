.PHONY: setup dev build test run

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
