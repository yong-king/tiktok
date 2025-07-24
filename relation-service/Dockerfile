FROM golang:1.24 AS builder


WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ralation-service ./cmd/ralation-service


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
COPY --from=builder /src/bin/ralation-service /app/ralation-service

# 把配置文件也拷贝进去镜像里
COPY configs/ /app/configs

WORKDIR /app

EXPOSE 8086
EXPOSE 9086
VOLUME /data/conf

CMD ["./ralation-service", "-conf", "/app/configs/config_doc.yaml"]
