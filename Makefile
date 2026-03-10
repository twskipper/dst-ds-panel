DOCKER_REPO = twskipper/dst-ds-panel
DST_REPO = twskipper/dst-ds-runtime

.PHONY: backend frontend docker-build docker-build-linux docker-build-panel docker-push docker-all dst-install dev build release tray app

# Build frontend, then embed into Go binary
build: frontend
	rm -rf backend/cmd/server/frontend
	cp -r frontend/dist backend/cmd/server/frontend
	cp frontend/public/world-settings.json backend/cmd/server/world-settings.json
	cd backend && go build -o dst-ds-panel ./cmd/server

# Cross-compile release binaries + macOS app bundle
release: frontend
	rm -rf backend/cmd/server/frontend dist
	cp -r frontend/dist backend/cmd/server/frontend
	cp frontend/public/world-settings.json backend/cmd/server/world-settings.json
	mkdir -p dist
	cd backend && GOOS=darwin GOARCH=arm64 go build -o ../dist/dst-ds-panel-darwin-arm64 ./cmd/server
	cd backend && GOOS=darwin GOARCH=amd64 go build -o ../dist/dst-ds-panel-darwin-amd64 ./cmd/server
	cd backend && GOOS=linux GOARCH=amd64 go build -o ../dist/dst-ds-panel-linux-amd64 ./cmd/server
	cd backend && go build -o dst-ds-panel-tray ./cmd/tray
	rm -rf "dist/DST DS Panel.app"
	mkdir -p "dist/DST DS Panel.app/Contents/MacOS" "dist/DST DS Panel.app/Contents/Resources"
	cp dist/dst-ds-panel-darwin-arm64 "dist/DST DS Panel.app/Contents/MacOS/dst-ds-panel"
	cp backend/dst-ds-panel-tray "dist/DST DS Panel.app/Contents/MacOS/dst-ds-panel-tray"
	cp frontend/public/icon.png "dist/DST DS Panel.app/Contents/Resources/icon.png"
	cp config.example.json "dist/DST DS Panel.app/Contents/MacOS/config.example.json"
	printf '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n\t<key>CFBundleExecutable</key>\n\t<string>dst-ds-panel-tray</string>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.dst-ds-panel</string>\n\t<key>CFBundleName</key>\n\t<string>DST DS Panel</string>\n\t<key>CFBundleVersion</key>\n\t<string>1.0.0</string>\n\t<key>LSUIElement</key>\n\t<true/>\n\t<key>CFBundleIconFile</key>\n\t<string>icon</string>\n</dict>\n</plist>' > "dist/DST DS Panel.app/Contents/Info.plist"
	cd dist && zip -r "DST.DS.Panel.app.zip" "DST DS Panel.app"
	@echo "Release artifacts in dist/:"
	@echo "  dst-ds-panel-darwin-arm64"
	@echo "  dst-ds-panel-darwin-amd64"
	@echo "  dst-ds-panel-linux-amd64"
	@echo "  DST.DS.Panel.app.zip"

# macOS menu bar tray app
tray:
	cd backend && go build -o dst-ds-panel-tray ./cmd/tray

# macOS .app bundle (tray + server)
app: build tray
	rm -rf "dist/DST DS Panel.app"
	mkdir -p "dist/DST DS Panel.app/Contents/MacOS"
	mkdir -p "dist/DST DS Panel.app/Contents/Resources"
	cp backend/dst-ds-panel "dist/DST DS Panel.app/Contents/MacOS/dst-ds-panel"
	cp backend/dst-ds-panel-tray "dist/DST DS Panel.app/Contents/MacOS/dst-ds-panel-tray"
	cp frontend/public/icon.png "dist/DST DS Panel.app/Contents/Resources/icon.png"
	cp config.example.json "dist/DST DS Panel.app/Contents/MacOS/config.example.json"
	echo '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0">\n<dict>\n\t<key>CFBundleExecutable</key>\n\t<string>dst-ds-panel-tray</string>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.dst-ds-panel</string>\n\t<key>CFBundleName</key>\n\t<string>DST DS Panel</string>\n\t<key>CFBundleVersion</key>\n\t<string>1.0.0</string>\n\t<key>LSUIElement</key>\n\t<true/>\n\t<key>CFBundleIconFile</key>\n\t<string>icon</string>\n</dict>\n</plist>' > "dist/DST DS Panel.app/Contents/Info.plist"
	@echo "macOS app bundle created: dist/DST DS Panel.app"

backend:
	cd backend && go build -o dst-ds-panel ./cmd/server

frontend:
	cd frontend && npm run build

# macOS: runtime-only image (DST mounted from host)
docker-build:
	docker build --platform linux/amd64 -f docker/Dockerfile.dst -t $(DST_REPO):macos docker/

# Linux amd64: self-contained image with SteamCMD
docker-build-linux:
	docker build --platform linux/amd64 -f docker/Dockerfile.linux -t $(DST_REPO):linux docker/

# Panel image (the web app itself)
docker-build-panel:
	docker build -f deploy/Dockerfile.panel -t $(DOCKER_REPO):latest .

# Push all images to Docker Hub
docker-push:
	docker push $(DST_REPO):macos
	docker push $(DST_REPO):linux
	docker push $(DOCKER_REPO):latest
	@echo "Pushed to Docker Hub:"
	@echo "  $(DOCKER_REPO):latest          (panel)"
	@echo "  $(DST_REPO):macos   (DST runtime for macOS)"
	@echo "  $(DST_REPO):linux   (DST with SteamCMD for Linux)"

# Build all Docker images and push to Docker Hub
docker-all: docker-build docker-build-linux docker-build-panel docker-push

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
