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
COPY --from=builder /src/start.sh /app/start.sh
COPY --from=builder /src/wait-for-it.sh /app/wait-for-it.sh

RUN dos2unix /app/start.sh /app/wait-for-it.sh && \
    chmod +x /app/start.sh /app/wait-for-it.sh


EXPOSE 8085
EXPOSE 9085
VOLUME /data/conf

CMD ["./start.sh"]
