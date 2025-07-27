FROM golang:1.24 AS builder

#COPY . /src
WORKDIR /src

# 复制 go.mod 和 go.sum 以便利用缓存拉取依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/video-service ./cmd/video-service


FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
# 复制编译好的二进制文件
COPY --from=builder /src/bin/video-service /app/video-service

# 把配置文件也拷贝进去镜像里
COPY configs/ /app/configs
COPY start.sh /app/start.sh
COPY wait-for-it.sh /app/wait-for-it.sh

RUN dos2unix /app/start.sh /app/wait-for-it.sh && \
    chmod +x /app/start.sh /app/wait-for-it.sh

EXPOSE 8082
EXPOSE 9082
VOLUME /data/conf


CMD ["./start.sh"]
