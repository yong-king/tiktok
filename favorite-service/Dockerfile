FROM golang:1.24 AS builder

#COPY . /src
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/favorite-service ./cmd/favorite-service

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /src/bin/favorite-service /app/favorite-service

COPY configs/ /app/configs


EXPOSE 8083
EXPOSE 9083
VOLUME /data/conf

CMD ["./favorite-service", "-conf", "/app/configs/config_doc.yaml"]
