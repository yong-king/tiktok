FROM golang:1.24 AS builder

#COPY . /src
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/feed-service ./cmd/feed-service

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /src/bin/feed-service /app/feed-service

COPY configs/ /app/configs


EXPOSE 8085
EXPOSE 9085
VOLUME /data/conf

CMD ["./feed-service", "-conf", "/app/configs/config_doc.yaml"]
