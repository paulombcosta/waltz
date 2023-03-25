alias start := run

run:
  go run .
build:
 go build
test:
  go generate ./... && go test ./...
