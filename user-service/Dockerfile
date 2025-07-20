FROM golang:1.24 AS builder

# 设置工作目录
#COPY . /src
WORKDIR /src

# 复制 go.mod 和 go.sum 以便利用缓存拉取依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/user-service ./cmd/user-service


#RUN GOPROXY=https://goproxy.cn make build



FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

#COPY --from=builder /src/bin /app
# 复制编译好的二进制文件
COPY --from=builder /src/bin/user-service /app/user-service

# 把配置文件也拷贝进去镜像里
COPY configs /app/configs

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./user-service", "-conf", "/app/configs/config.yaml"]
