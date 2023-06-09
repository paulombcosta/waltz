FROM golang:1.20.4 AS builder

COPY . /build

WORKDIR /build

RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o waltz

FROM alpine:3.16.2

WORKDIR /app

COPY --from=builder /build/waltz /app

ENTRYPOINT [ "/app/waltz" ]
