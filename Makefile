.PHONY: test
test:
	GCP_API_KEY=1234 go test ./...

.PHONY: format
format:
	gofumpt -s -w .
