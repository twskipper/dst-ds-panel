.PHONY: backend frontend docker-build docker-build-linux dst-install dev build release

# Build frontend, then embed into Go binary
build: frontend
	rm -rf backend/cmd/server/frontend
	cp -r frontend/dist backend/cmd/server/frontend
	cp frontend/public/world-settings.json backend/cmd/server/world-settings.json
	cd backend && go build -o dst-ds-panel ./cmd/server

# Cross-compile release binaries for all platforms
release: frontend
	rm -rf backend/cmd/server/frontend dist
	cp -r frontend/dist backend/cmd/server/frontend
	cp frontend/public/world-settings.json backend/cmd/server/world-settings.json
	mkdir -p dist
	cd backend && GOOS=darwin GOARCH=arm64 go build -o ../dist/dst-ds-panel-darwin-arm64 ./cmd/server
	cd backend && GOOS=darwin GOARCH=amd64 go build -o ../dist/dst-ds-panel-darwin-amd64 ./cmd/server
	cd backend && GOOS=linux GOARCH=amd64 go build -o ../dist/dst-ds-panel-linux-amd64 ./cmd/server
	@echo "Release binaries in dist/"

backend:
	cd backend && go build -o dst-ds-panel ./cmd/server

frontend:
	cd frontend && npm run build

# macOS (Apple Silicon): runtime-only image, DST mounted from host
docker-build:
	docker build --platform linux/amd64 -f docker/Dockerfile.dst -t dst-server:latest docker/

# Linux amd64: self-contained image with SteamCMD, installs DST on first run
docker-build-linux:
	docker build -f docker/Dockerfile.linux -t dst-server:latest docker/

# macOS only: download/update DST server via DepotDownloader
dst-install:
	./scripts/install-dst.sh

dst-update: dst-install

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev:
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals"
