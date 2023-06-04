alias start := run

run:
  PROFILE=dev go run .
build:
 go build
test:
  go generate ./... && go test ./...
lint:
  golangci-lint run ./...
run-search:
  docker-compose stop
  docker-compose build
  docker-compose up -d
  docker logs -f search
