# Manjing AI Go

后端最小骨架，包含：Gin、GORM、JWT、Redis 黑名单、Swagger、迁移工具。

## 快速开始

1) 修改 `config.yaml`
2) 迁移数据库：
   - `go run ./cmd/migrate up`
3) 启动 API：
   - `go run ./cmd/api`
4) 安装 swag：
   - `go install github.com/swaggo/swag/cmd/swag@v1.16.3`
5) 生成 Swagger：
   - `swag init -g cmd/api/main.go -o swagger`
6) Swagger：
   - `/swagger/index.html`

## 目录结构

```
cmd/            入口
config/         配置加载
internal/       业务代码
pkg/            公共库
migrations/     数据库迁移
swagger/        Swagger 文档
```

## JWT 续期

当 Token 剩余有效期 <= `renew_threshold_days` 时，服务端自动签发新 Token：
- 响应头只返回 `X-Token`
- 前端收到后替换本地 Token
