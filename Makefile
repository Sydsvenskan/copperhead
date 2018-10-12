.PHONY: cover
cover:
	go test -race -coverprofile=/tmp/coverage.out \
		&& go tool cover -html=/tmp/coverage.out
