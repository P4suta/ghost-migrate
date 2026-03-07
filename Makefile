.PHONY: build test lint clean update-golden

build:
	go build -o bin/ghost-migrate ./cmd/ghost-migrate

test:
	go test -v -race -count=1 ./...

lint:
	go vet ./...
	@if [ -n "$$(gofmt -l .)" ]; then echo "gofmt found issues:"; gofmt -l .; exit 1; fi

clean:
	rm -rf bin/ output/

update-golden:
	UPDATE_GOLDEN=1 go test ./internal/adapter/htmlconv/...
