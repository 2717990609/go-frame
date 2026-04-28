# Go Backend Framework

**企业级通用Go后端框架** - 快速启动、配置驱动、插件化扩展

## 🚀 核心特性

- **🔧 配置驱动**: 100%环境变量支持，${VAR:default}语法
- **🗄️ 多数据库支持**: MySQL、PostgreSQL、SQLite、MongoDB
- **🔌 插件化架构**: 可插拔中间件，热加载支持
- **🛡️ 企业级安全**: 参数校验、防抖、签名验证、链路追踪
- **📊 可观测性**: 结构化日志、健康检查、Metrics监控
- **🏗️ 标准架构**: Handler→Service→Repository→Model分层

## 📁 标准目录结构

```
go-backend-framework/
├── cmd/
│   └── server/
│       └── main.go              # 启动入口，仅负责组装
├── internal/                     # 业务代码，对外不可见
│   ├── handler/                  # HTTP层，参数校验、调用Service
│   ├── service/                  # 业务逻辑层
│   ├── repository/               # 数据访问层，仅与DB交互
│   ├── model/                    # PO/DO定义
│   └── dto/                      # Request/Response/VO
├── pkg/                          # 可复用公共包
│   ├── database/                 # 数据库抽象层
│   ├── config/                   # 配置引擎
│   ├── middleware/               # 认证、限流、RequestID等
│   ├── logger/                   # 日志封装，支持trace_id
│   ├── validator/                # 参数校验
│   ├── response/                 # 统一响应封装
│   ├── debounce/                 # 防抖工具
│   └── framework/                # 基础设施
├── api/                          # 路由组装
├── config/                       # 配置文件
├── docs/                         # Swagger文档
├── migrations/                   # 数据库迁移脚本
└── plugins/                      # 插件目录
```

## 🎯 快速开始

### 前置依赖

- Go 1.20+
- 数据库 (MySQL 8.0+ / PostgreSQL 12+ / Redis 6+)

### 启动步骤

1. **克隆并进入项目**
   ```bash
   cd go-backend-framework
   ```

2. **配置数据库和环境**
   ```bash
   cp config/config.example.yaml config/config.dev.yaml
   # 修改配置文件，或使用环境变量：
   export DB_HOST=localhost
   export DB_NAME=myapp
   export DB_PASSWORD=mypassword
   export REDIS_ADDR=localhost:6379
   ```

3. **创建数据库**
   ```sql
   CREATE DATABASE IF NOT EXISTS app_default CHARACTER SET utf8mb4;
   ```

4. **安装依赖并运行**
   ```bash
   go mod download
   go run ./cmd/server/main.go -config config/config.dev.yaml

   ```

5. **验证服务**
   ```bash
   # 存活检查
   curl http://localhost:8080/health
   # 就绪检查
   curl http://localhost:8080/ready
   # Swagger文档
   open http://localhost:8080/swagger/index.html
   ```

## 🔧 核心能力

### 配置系统
```yaml
# 支持 ${VAR:default} 语法
database:
  driver: "${DB_DRIVER:mysql}"
  host: "${DB_HOST:localhost}"
  port: "${DB_PORT:3306}"
  database: "${DB_NAME:app_default}"
```

### 数据库抽象层
```go
// 统一的数据库接口
type DatabaseProvider interface {
    Find(ctx context.Context, dest any, query Query) error
    Create(ctx context.Context, model any) error
    Transaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
    // ...
}

// 支持多种数据库
db, err := database.New(database.Config{
    Driver: "mysql",  // postgresql / sqlite / mongodb
    Host:   "localhost",
    // ...
})
```

### 插件系统
```go
// 标准插件接口
type Plugin interface {
    Name() string
    Init(ctx context.Context, config map[string]interface{}) error
    Routes() []Route
    Middlewares() []Middleware
}

// 配置化加载
plugins.enabled: "swagger,metrics,tracing,discovery"
```

### 统一响应格式
```go
// 成功
response.Success(data)

// 错误（禁止返回nil data）
response.Error(response.CodeParamError, "参数错误")
```

### 错误码体系
| 范围 | 含义 |
|------|------|
| 200 | 成功 |
| 4000-4099 | 客户端参数错误 |
| 4100-4199 | 业务逻辑错误 |
| 4200-4299 | 权限/状态错误 |
| 5000-5099 | 服务端错误 |
| 6000-6099 | 第三方服务错误 |

### 中间件链（标准顺序）
1. **Recovery** - Panic捕获
2. **RequestID** - trace_id生成与透传
3. **Logger** - 请求日志
4. **Timeout** - 请求超时（默认30s）
5. **Auth** - JWT认证
6. **RateLimit** - 限流
7. **CORS** - 跨域

### 防抖工具
```go
ok, err := debounce.Apply(ctx, redis, "user:123:withdraw", 10*time.Second)
if !ok {
    return errors.New("操作进行中，请勿重复提交")
}
```

## 🏗️ 开发新模块

按分层规范新增模块：

1. **model** - 定义PO结构体
2. **repository** - 实现数据访问
3. **dto** - 定义Request/Response/VO
4. **service** - 实现业务逻辑
5. **handler** - 实现HTTP接口，添加Swagger注释
6. **api/router.go** - 注册路由

## 🔐 生产部署

### 环境变量配置
```bash
# 敏感配置环境变量注入
export DB_PASSWORD="your-production-password"
export REDIS_PASSWORD="your-redis-password"
export JWT_SECRET="your-jwt-secret"
export SIGNATURE_KEY="your-signature-key"
export SERVER_PORT=80
```

### 健康检查
- **存活检查**: `/health`（无依赖检查）
- **就绪检查**: `/ready`（含数据库、Redis检查）

### 监控支持
```yaml
monitoring:
  prometheus:
    enabled: true
    path: /metrics
  tracing:
    enabled: true
    jaeger:
      endpoint: http://jaeger:14268/api/traces
```

## 🚨 安全特性

- **参数校验**: 自动参数验证和格式化
- **SQL注入防护**: 预处理语句 + ORM安全
- **签名验证**: 敏感接口请求验签
- **防重放**: Nonce时间窗口 + 防抖机制
- **日志安全**: 敏感字段自动脱敏

## 📚 更多文档

- [开发规范](./docs/开发规范.md) - 详细的开发指南
- [配置参考](./docs/配置参考.md) - 完整的配置选项说明
- [插件开发](./docs/插件开发.md) - 如何开发自定义插件
- [部署手册](./docs/部署手册.md) - 生产环境部署指南

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 License

内部项目，保密

---

**Go Backend Framework** - 让企业级Go后端开发更简单！