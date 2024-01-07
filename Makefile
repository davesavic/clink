#Makefile

# Test
test:
	go test -v -failfast -count=1 -cover -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func coverage.out

.phony: test
