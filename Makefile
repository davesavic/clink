#Makefile

# Test
test:
	go test -v -cover -coverprofile=coverage.out ./...

.phony: test