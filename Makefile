test:
	go test ./... -v -count=1

test-race:
	go test ./... -race -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

clean:
	rm -f coverage.out coverage.html

.PHONY: test test-race test-coverage lint clean
