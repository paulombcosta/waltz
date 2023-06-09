alias start := run

run:
  PROFILE=dev air
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
docker-build:
  docker build -t paulombcosta/waltz:latest .
docker-run:
  docker run -it -p "8080:8080" paulombcosta/waltz:latest
