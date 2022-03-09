.PHONY: build format test run clean

format:
	go fmt ./...

test:
	go test -v ./...

build-collector:
	go build -v -o ./bin/collector ./cmd/collector/

build-preprocessor:
	go build -v -o ./bin/preprocessor ./cmd/preprocessor/

build-archiver:
	go build -v -o ./bin/archiver ./cmd/archiver/

run-collector:
	go run ./cmd/collector/main.go

run-preprocessor:
	go run ./cmd/preprocessor/main.go

run-archiver:
	go run ./cmd/archiver/main.go

clean:
	rm -f ./bin/