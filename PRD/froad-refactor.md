# Go-Backend-Frame 框架通用化改造 PRD

## 0. 元信息
- **文档状态**: Active
- **项目版本**: v1.0.0 (重构中)
- **更新时间**: 2026-03-02
- **改造模式**: 激进式重构
- **可信度说明**: S1基于现有系统，S3为架构设计

---

## 1. 业务背景（为什么存在）

### 1.1 现状分析（S1）
【事实】当前框架存在以下问题：
- 模块名硬编码为"fire-mirage" [source: go.mod]
- 数据库名、配置项硬编码 [source: config/config.dev.yaml]
- 业务实体耦合度高（如User.Spark星火余额） [source: internal/model/user.go]
- 技术栈过时（Redis v8、Gorm等） [source: go.mod依赖]

### 1.2 目标定位（S3）
【合理推测】建立企业级通用框架，支持：
- 多项目快速启动
- 多数据库类型切换  
- 配置驱动架构
- 插件化扩展

---

## 2. 业务对象与业务规则（系统如何运转）

### 2.1 核心对象设计（S3）

```yaml
core_objects:
- name: FrameworkEngine
  description: 框架引擎，负责组件组装和启动
  
- name: DatabaseProvider  
  description: 数据库提供者抽象接口
  
- name: ConfigEngine
  description: 配置引擎，支持环境变量注入
  
- name: PluginRegistry
  description: 插件注册表，管理插件生命周期
```

### 2.2 业务规则（S2+S3）

【S2 现有规则】：
- 保持错误码体系不变 [source: pkg/response/response.go]
- 保持trace_id链路追踪 [source: docs/后端开发规范.md]
- 保持分层架构原则 [source: README.md]

【S3 新增规则】：
- 所有配置项必须支持环境变量覆盖
- 数据库切换必须在配置层面完成
- 插件必须实现标准Plugin接口

---

## 3. 功能与需求（我们做了什么）

### 3.1 重构需求清单

#### Phase 1: 基础重构（1-2周）
【S3 设计需求】：

```
3.1.1 模块解耦任务
- 重命名 module: fire-mirage → go-backend-framework
- 移除所有硬编码项目标识
- 建立通用项目命名规范

3.1.2 配置系统重构  
- 实现 ${VAR:default} 语法支持
- 支持嵌套配置结构
- 支持多环境配置管理

3.1.3 数据库抽象层设计
- 定义 DatabaseProvider 接口
- 实现 MySQL 适配器
- 提供迁移兼容层
```

#### Phase 2: 技术栈升级（2-3周）
【S3 技术需求】：

```
3.2.1 ORM替换
- 评估: Gorm → SQLx + QueryBuilder
- 性能测试: 查询性能基准测试
- 迁移工具: 自动化模型转换工具

3.2.2 Web框架升级
- 候选: Echo/Fiber/自建框架
- 评估维度: 性能、生态、学习曲线
- 兼容性: 保持现有中间件接口
```

#### Phase 3: 插件生态（3-4周）  
【S3 生态需求】：

```
3.3.1 插件系统
- Plugin接口设计
- 插件注册机制
- 热插拔支持

3.3.2 企业插件
- Swagger文档插件
- Metrics监控插件  
- Tracing链路追踪插件
- ServiceDiscovery服务发现插件
```

---

## 4. 历史决策与妥协（为什么这么做）

### 4.1 已知历史决策（S1/S2）

【S1 文档确认的决策】：
- 选择Go 1.20+作为基础语言 [source: go.mod]
- 采用Gin+Gorm+Redis技术栈 [source: go.mod依赖]
- 制定详细开发规范v1.8 [source: docs/后端开发规范.md]

### 4.2 重构决策说明（S3）

**为什么选择配置驱动而不是代码生成？**
- 更灵活的项目定制需求
- 避免代码生成带来的维护成本
- 更好的运行时可调整性

**为什么保留现有错误码体系？**
- 企业级项目稳定性优先
- 避免大规模业务代码改动
- 兼容现有客户端对接

---

## 5. 风险与测试关注点

### 5.1 风险评估（S3）

```yaml
high_risks:
- id: R-001  
  description: 破坏现有业务系统稳定性
  mitigation: 渐进式迁移、兼容性适配器
  
- id: R-002
  description: 配置复杂度暴增
  mitigation: 提供默认配置、配置校验工具
  
medium_risks:
- id: R-003
  description: 性能回归
  mitigation: 基准测试、性能监控
```

### 5.2 测试策略（S3）

- **单元测试**: 每个抽象层必须有100%覆盖
- **集成测试**: 数据库切换场景测试
- **性能测试**: 与现有框架性能对比
- **兼容性测试**: 现有API接口兼容性验证

---

## 6. 变更与演进记录

→ 详细变更请查阅 `./CHANGELOG.md`

---

## 7. 附录：技术实施细节

### 7.1 配置系统设计（S3）

```go
type ConfigEngine struct {
    data map[string]interface{}
}

func (c *ConfigEngine) Get(key string) interface{}
func (c *ConfigEngine) GetString(key string, defaultValue string) string  
func (c *ConfigEngine) GetStruct(key string, dest interface{}) error
```

### 7.2 数据库抽象接口（S3）

```go
type DatabaseProvider interface {
    Find(ctx context.Context, dest any, query Query) error
    Create(ctx context.Context, model any) error
    Transaction(ctx context.Context, fn func(ctx context.Context, Tx) error) error
}
```

### 7.3 插件接口设计（S3）

```go
type Plugin interface {
    Name() string
    Init(ctx context.Context, config map[string]interface{}) error
    Routes() []Route
    Middlewares() []Middleware
}
```

---

**⚡ AI自检清单**：
- ✅ 未引入未确认的新概念（均标注为S3）
- ✅ 保留了历史决策说明
- ✅ 能够解释"为何如此设计"  
- ✅ 新人可基于此理解重构计划
- ✅ 遵循PRD标准结构
- ✅ 风险评估和缓解策略明确

**下一步**: 开始Phase 1具体实施，创建技术文档和配置系统原型