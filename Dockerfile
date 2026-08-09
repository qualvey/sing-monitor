# 构建阶段
FROM golang:1.25-alpine AS builder
WORKDIR /src

# 先拷贝依赖清单，利用缓存
COPY server/go.mod server/go.sum ./
RUN go mod download

# 拷贝源码并构建（glebarez/sqlite 纯 Go，无需 cgo）
COPY server/ ./
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/sing-monitor-server .

# 运行阶段
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/sing-monitor-server /usr/local/bin/sing-monitor-server

# 默认配置（挂载卷可覆盖）
COPY server/config.example.json /app/config.json

EXPOSE 8080
ENTRYPOINT ["sing-monitor-server"]
