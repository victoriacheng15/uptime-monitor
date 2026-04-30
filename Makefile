.PHONY: format test coverage update lambda-build setup-tailwind web-build

format:
	gofmt -w ./cmd ./internal

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

update:
	go get -u ./...
	go mod tidy

lambda-build:
	go build -o bin/lambda ./cmd/lambda

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
