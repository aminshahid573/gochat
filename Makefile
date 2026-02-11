build:
	@go build -o bin/gochat .

run: build
	@./bin/gochat

