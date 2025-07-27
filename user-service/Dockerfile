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



#FROM debian:stable-slim
#
#RUN apt-get update && apt-get install -y --no-install-recommends \
#		ca-certificates  \
#        netbase \
#        && rm -rf /var/lib/apt/lists/ \
#        && apt-get autoremove -y && apt-get autoclean -y

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

#COPY --from=builder /src/bin /app
# 复制编译好的二进制文件
COPY --from=builder /src/bin/user-service /app/user-service

# 把配置文件也拷贝进去镜像里
COPY configs /app/configs
COPY start.sh /app/start.sh
COPY wait-for-it.sh /app/wait-for-it.sh


# ✅ 转换换行符（Windows ➜ Linux），并添加执行权限
RUN dos2unix /app/start.sh /app/wait-for-it.sh && \
    chmod +x /app/start.sh /app/wait-for-it.sh

WORKDIR /app

EXPOSE 8081
EXPOSE 9081
VOLUME /data/conf

CMD ["./start.sh"]
