# Fire Mirage Makefile

.PHONY: run build swag lint test dev

# 开发环境运行
run:
	go run ./cmd/server -config config/config.dev.yaml

# 构建
build:
	go build -o bin/server ./cmd/server

# 生成 Swagger 文档
swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 静态检查
lint:
	golangci-lint run

# 测试
test:
	go test ./...

# 开发：启动服务（需本地 MySQL、Redis）
dev: build
	./bin/server -config config/config.dev.yaml
