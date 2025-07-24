FROM golang:1.24 AS builder

#COPY . /src
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/job-service ./cmd/job-service

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /src/bin/job-service /app/job-service

COPY configs/ /app/configs


EXPOSE 8087
EXPOSE 9087
VOLUME /data/conf

CMD ["./job-service", "-conf", "/app/configs/config_doc.yaml"]
