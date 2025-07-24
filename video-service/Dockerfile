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
#
#RUN apt-get update && apt-get install -y --no-install-recommends \
#		ca-certificates  \
#        netbase \
#        && rm -rf /var/lib/apt/lists/ \
#        && apt-get autoremove -y && apt-get autoclean -y

WORKDIR /app
# 复制编译好的二进制文件
COPY --from=builder /src/bin/video-service /app/video-service

# 把配置文件也拷贝进去镜像里
COPY configs/ /app/configs

WORKDIR /app

EXPOSE 8082
EXPOSE 9082
VOLUME /data/conf

CMD ["./video-service", "-conf", "/app/configs/config_doc.yaml"]
