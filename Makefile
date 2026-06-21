.PHONY: format test test-bdd test-all coverage update lambda-build lambda-package setup-tailwind web-build

format:
	gofmt -w ./cmd ./internal

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
