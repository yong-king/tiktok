FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/job-service ./cmd/job-service


FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /src/bin/job-service /app/job-service

COPY configs/ /app/configs

COPY start.sh /app/start.sh
COPY wait-for-it.sh /app/wait-for-it.sh

RUN dos2unix /app/start.sh /app/wait-for-it.sh && \
    chmod +x /app/start.sh /app/wait-for-it.sh


EXPOSE 8087
EXPOSE 9087
VOLUME /data/conf


CMD ["./start.sh"]
