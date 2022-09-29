.PHONY: quality
quality:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	golangci-lint run

.PHONY: build
build:
	goreleaser release --snapshot --rm-dist