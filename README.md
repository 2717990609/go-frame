# Fire Mirage（星火现梦）后端基础框架

基于《星火现梦》后端开发规范 v1.8 构建的 Go 后端基础框架，适用于多项目复用。

## 📁 项目结构

```
fire-mirage/
├── cmd/
│   └── server/
│       └── main.go              # 入口，仅负责组装与启动
├── internal/                     # 业务代码，对外不可见
│   ├── handler/                  # HTTP 层，参数校验、调用 Service
│   ├── service/                  # 业务逻辑层
│   ├── repository/               # 数据访问层，仅与 DB 交互
│   ├── model/                    # PO/DO 定义
│   └── dto/                      # Request/Response/VO 结构体
├── pkg/                          # 可复用公共包
│   ├── logger/                  # 日志封装，支持 trace_id 全链路透传
│   ├── middleware/              # 认证、限流、RequestID、Recovery 等
│   ├── validator/               # 参数校验
│   ├── response/                # 统一响应封装
│   ├── debounce/                # 防抖工具
│   └── framework/               # 数据库、Redis 等基础设施
├── api/                         # 路由组装
├── config/                      # 配置文件
├── docs/                        # Swagger 文档
├── migrations/                  # 数据库迁移脚本
└── docs/                        # 规范文档
```

## 🚀 快速开始

### 前置依赖

- Go 1.20+
- MySQL 8.0+
- Redis 6+

### 启动步骤

1. **克隆并进入项目**
   ```bash
   cd fire-mirage
   ```

2. **复制配置文件**
   ```bash
   cp config/config.example.yaml config/config.dev.yaml
   # 按需修改 MySQL、Redis 等配置
   ```

3. **创建数据库**
   ```sql
   CREATE DATABASE fire_mirage CHARACTER SET utf8mb4;
   ```

4. **安装依赖并运行**
   ```bash
   go mod download
   go run ./cmd/server -config config/config.dev.yaml
   ```

5. **验证**
   - 存活检查: `curl http://localhost:8080/health`
   - 就绪检查: `curl http://localhost:8080/ready`
   - Swagger: `http://localhost:8080/swagger/index.html`

## 🔧 核心能力

### 统一响应格式

```go
// 成功
response.Success(data)

// 错误（禁止返回 nil data）
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

### 日志规范（trace_id 透传）

```go
logger.C(ctx).Info("用户登录成功",
    zap.Int64("user_id", userID),
    zap.String("login_type", "wechat"),
)
```

### 中间件链（顺序固定）

1. Recovery - panic 捕获
2. RequestID - trace_id 生成与透传
3. Logger - 请求日志
4. Timeout - 请求超时（默认 30s）
5. CORS - 跨域
6. RateLimit - 限流

### 防抖工具

```go
ok, err := debounce.Apply(ctx, redis, "user:123:withdraw", 10*time.Second)
if !ok {
    return errors.New("操作进行中，请勿重复提交")
}
```

### 请求验签

敏感接口（提现、回调等）需验签，详见 [docs/参数加密验签方案.md](docs/参数加密验签方案.md)。

```go
// 客户端生成签名
sign, ts, nonce := signature.GenerateWithNonce(params, secret)
// Header: X-Signature, X-Timestamp, X-Nonce

// 服务端：中间件自动验签
signed := v1.Group("/signed")
signed.Use(middleware.Signature(cfg.Signature, rdb))
```

## 📝 开发新模块

按分层规范新增模块：

1. **model** - 定义 PO 结构体
2. **repository** - 实现数据访问
3. **dto** - 定义 Request/Response/VO
4. **service** - 实现业务逻辑
5. **handler** - 实现 HTTP 接口，添加 Swagger 注释
6. **api/router.go** - 注册路由

## 🔐 生产部署

- 敏感配置通过环境变量注入：`MYSQL_PASSWORD`、`REDIS_PASSWORD`、`JWT_SECRET`
- 生产配置 `config.prod.yaml` 禁止提交仓库
- 健康检查：`/health`（存活）、`/ready`（就绪，含 MySQL/Redis 检查）

## 📚 规范文档

详见 [docs/后端开发规范.md](docs/后端开发规范.md)

## 📄 License

内部项目，保密
