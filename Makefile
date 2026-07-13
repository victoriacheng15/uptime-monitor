.PHONY: format format-md lint-md test test-bdd test-all coverage update lambda-build lambda-package setup-tailwind web-build dev-build dev-run dev-clean

format:
	gofmt -w ./cmd ./internal

format-md:
	npx markdownlint-cli --fix "**/*.md"

lint-md:
	@echo "Linting Markdown files..."
	npx markdownlint-cli "**/*.md"

test:
	go test -v ./cmd/... ./internal/...

test-bdd:
	go test -v ./e2e/...

test-all:
	@$(MAKE) test
	@$(MAKE) test-bdd

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

update:
	go get -u ./...
	go mod tidy

lambda-build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/bootstrap ./cmd/lambda

lambda-package: lambda-build
	cd bin && zip -q lambda.zip bootstrap

setup-tailwind:
	@echo "Downloading tailwind css cli..." && \
	curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss && \
	chmod +x tailwindcss

web-build: setup-tailwind
	@echo "Generating static site..." && \
	rm -rf dist && \
	mkdir -p dist && \
	go run ./cmd/web && \
	./tailwindcss -i ./internal/web/templates/styles.css -o ./dist/styles.css --minify && \
	rm tailwindcss

dev-build:
	podman build -t uptime-monitor-dev .

dev-run:
	podman run -it --rm \
		-p 8080:8080 \
		-v "$(PWD):/app:Z" \
		-e MONITOR_TARGETS="https://google.com,https://github.com" \
		uptime-monitor-dev

dev-clean:
	podman rmi uptime-monitor-dev
