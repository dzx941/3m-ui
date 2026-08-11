APP_NAME := 3m-ui
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DIST_DIR := dist

.PHONY: build build-linux release clean

build:
	cd $(FRONTEND_DIR) && pnpm build
	rm -rf $(BACKEND_DIR)/cmd/server/web/dist
	cp -R $(FRONTEND_DIR)/dist $(BACKEND_DIR)/cmd/server/web/dist
	cd $(BACKEND_DIR) && go build -trimpath -ldflags "-s -w" -o ../$(DIST_DIR)/$(APP_NAME) ./cmd/server

build-linux:
	mkdir -p $(DIST_DIR)
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o ../$(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o ../$(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o ../$(DIST_DIR)/$(APP_NAME)-linux-armv7 ./cmd/server

release: clean build build-linux
	cp scripts/install.sh scripts/update.sh scripts/uninstall.sh $(DIST_DIR)/

clean:
	rm -rf $(DIST_DIR)
