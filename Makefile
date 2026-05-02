.PHONY: help setup dev build test run \
        docker-build docker-run docker-stop \
        compose-up compose-down compose-logs \
        bootstrap release

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
	@echo ""
	@echo "Release"
	@echo "  release        Bump version, commit, tag, and push"

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

# ---- release ----

release:
	@CURRENT=$$(grep 'const Version' internal/version/version.go | sed 's/.*"\(.*\)".*/\1/'); \
	echo "Current version: $$CURRENT"; \
	echo ""; \
	\
	BASE=$$(echo $$CURRENT | sed 's/-.*//'); \
	MAJOR=$$(echo $$BASE | cut -d. -f1 | tr -d 'v'); \
	MINOR=$$(echo $$BASE | cut -d. -f2); \
	PATCH=$$(echo $$BASE | cut -d. -f3); \
	PRESUFFIX=$$(echo $$CURRENT | grep -o '\-.*' | tr -d '-'); \
	PRETYPE=$$(echo $$PRESUFFIX | cut -d. -f1); \
	PRENUM=$$(echo $$PRESUFFIX | grep -o '\.[0-9]*$$' | tr -d '.'); \
	\
	NEXTPRENUM=$$(( $${PRENUM:-0} + 1 )); \
	case "$$PRETYPE" in \
		alpha) NEXTSTAGE="v$$MAJOR.$$MINOR.$$PATCH-beta" ;; \
		beta)  NEXTSTAGE="v$$MAJOR.$$MINOR.$$PATCH-rc" ;; \
		rc)    NEXTSTAGE="v$$MAJOR.$$MINOR.$$PATCH" ;; \
		*)     NEXTSTAGE="" ;; \
	esac; \
	V_PRE_NUM="v$$MAJOR.$$MINOR.$$PATCH-$$PRETYPE.$$NEXTPRENUM"; \
	V_PRE_STAGE="$$NEXTSTAGE"; \
	V_PATCH="v$$MAJOR.$$MINOR.$$((PATCH+1))"; \
	V_MINOR="v$$MAJOR.$$((MINOR+1)).0"; \
	V_MAJOR="v$$((MAJOR+1)).0.0"; \
	\
	echo "What do you want to bump?"; \
	if [ -n "$$PRETYPE" ]; then \
		echo "  1) pre-release number  [$$V_PRE_NUM]"; \
		if [ -n "$$V_PRE_STAGE" ]; then \
			echo "  2) pre-release stage   [$$V_PRE_STAGE]"; \
		else \
			echo "  2) stable release      [$$V_PRE_STAGE]"; \
		fi; \
	fi; \
	echo "  3) patch               [$$V_PATCH]"; \
	echo "  4) minor               [$$V_MINOR]"; \
	echo "  5) major               [$$V_MAJOR]"; \
	echo ""; \
	printf "Choice: "; read CHOICE; \
	\
	case "$$CHOICE" in \
	1) [ -z "$$PRETYPE" ] && { echo "No pre-release suffix to bump."; exit 1; }; NEW="$$V_PRE_NUM" ;; \
	2) [ -z "$$PRETYPE" ] && { echo "Already a stable release."; exit 1; }; NEW="$$V_PRE_STAGE" ;; \
	3) NEW="$$V_PATCH" ;; \
	4) NEW="$$V_MINOR" ;; \
	5) NEW="$$V_MAJOR" ;; \
	*) echo "Invalid choice."; exit 1 ;; \
	esac; \
	\
	echo ""; \
	echo "  $$CURRENT  →  $$NEW"; \
	printf "Confirm? [y/N]: "; read OK; \
	if [ "$$OK" != "y" ] && [ "$$OK" != "Y" ]; then echo "Aborted."; exit 1; fi; \
	\
	sed -i.bak "s|$$CURRENT|$$NEW|g" internal/version/version.go && rm internal/version/version.go.bak; \
	git add internal/version/version.go; \
	git commit -m "chore: release $$NEW"; \
	git tag $$NEW; \
	git push && git push origin $$NEW; \
	echo ""; \
	echo "Released $$NEW — GitHub Actions will build and push the Docker image."

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
