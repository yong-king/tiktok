FROM golang:1.24 AS builder

#COPY . /src
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/comment-service ./cmd/comment-service

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /src/bin/comment-service /app/comment-service

COPY configs/ /app/configs

COPY start.sh /app/start.sh
COPY wait-for-it.sh /app/wait-for-it.sh

RUN dos2unix /app/start.sh /app/wait-for-it.sh && \
    chmod +x /app/start.sh /app/wait-for-it.sh

EXPOSE 8084
EXPOSE 9084
VOLUME /data/conf

#CMD ["./comment-service", "-conf", "/app/configs/config_doc.yaml"]

CMD ["./start.sh"]