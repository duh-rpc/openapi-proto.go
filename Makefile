.PHONY: test lint tidy fmt coverage ci clean

test:
	go test -v ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy && git diff --exit-code

fmt:
	go fmt ./... && git diff --exit-code

ci: tidy fmt lint test
	@echo
	@echo "\033[32mEVERYTHING PASSED!\033[0m"

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -f coverage.out coverage.html
	go clean
