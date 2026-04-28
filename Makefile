# Fire Mirage Makefile

.PHONY: run build swag lint test test-short test-plugin ci dev

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

# 快速测试（跳过重依赖集成场景）
test-short:
	go test -short ./...

# 插件模块测试
test-plugin:
	go test ./pkg/plugin/...

# CI 本地对齐
ci: lint test-short test-plugin build

# 开发：启动服务（需本地 MySQL、Redis）
dev: build
	./bin/server -config config/config.dev.yaml
