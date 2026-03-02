# v1.0.0 变更日志

## 概述
- 重构目标：将 fire-mirage 升级为通用 go-backend-framework
- 重构模式：激进式重构，允许技术栈替换
- 配置策略：100%配置驱动，零硬编码
- 数据库策略：支持多类型切换

## 文件变更清单
| 文件 | 变更类型 | 说明 |
|------|----------|------|
| go.mod | 重构 | 模块名从 fire-mirage 改为 go-backend-framework |
| config/config.dev.yaml | 重构 | 支持 ${VAR:default} 语法 |
| pkg/database/ | 新增 | 数据库抽象层设计 |
| pkg/config/ | 重构 | 配置引擎重构 |
| docs/ | 重构 | 文档结构通用化 |
| README.md | 重构 | 项目描述通用化 |

## 详细变更

### 日期: 2026-03-02

#### go.mod (重构)
```diff
- module fire-mirage
+ module go-backend-framework
```
**理由**: 移除项目特定命名，支持多项目复用

#### config/config.dev.yaml (重构)
```yaml
# 新增环境变量支持
mysql:
  database: "${DB_NAME:app_default}"
  host: "${DB_HOST:localhost}"
```
**理由**: 支持不同环境配置，提高通用性

#### pkg/database/interface.go (新增)
```go
// 数据库提供者抽象接口
type DatabaseProvider interface {
    Find(ctx context.Context, dest any, query Query) error
    Create(ctx context.Context, model any) error
    Transaction(ctx context.Context, fn func(ctx context.Context, Tx) error) error
}
```
**理由**: 为多数据库支持提供抽象层

#### pkg/config/engine.go (重构)
```go
// 新版配置引擎
type ConfigEngine struct {
    data map[string]interface{}
}

func (c *ConfigEngine) GetString(key string, defaultValue string) string
```
**理由**: 支持环境变量解析和嵌套配置

#### README.md (重构)
- 移除"星火现梦"项目特定描述
- 改为通用企业级Go框架文档
- 保留技术架构说明

### Phase 1 实施细节 (2026-03-02 ~ 2026-03-16)

#### Week 1: 基础解耦
- 完成模块重命名
- 移除硬编码业务标识
- 建立配置系统原型

#### Week 2: 数据库抽象  
- 实现MySQL适配器
- 设计查询构建器
- 开发兼容性适配器

### 技术债务清理
- 移除Redis v8，升级到v9
- 清理无效的错误处理
- 统一日志格式

### 性能优化
- 配置解析性能提升
- 数据库连接池优化
- 中间件加载性能改进

## 风险缓解措施

### 兼容性保障
```go
// 保留现有API签名
type LegacyGormAdapter struct {
    newProvider DatabaseProvider
}

func (l *LegacyGormAdapter) Find(dest interface{}, conds ...interface{}) error {
    // 内部使用新接口
    return l.newProvider.Find(ctx, dest, convertLegacyQuery(conds))
}
```

### 渐进式迁移
- 提供配置开关控制新旧组件
- 保持现有错误码体系
- 保留原有中间件接口签名

## 测试验证

### 单元测试覆盖
- 配置引擎功能测试
- 数据库抽象层测试
- 兼容性适配器测试

### 集成测试场景
- 多数据库切换测试
- 配置热加载测试
- 现有业务接口回归测试

### 性能基准测试
- 配置解析性能对比
- 数据库查询性能对比  
- HTTP请求处理性能对比

## 后续计划

### Phase 2 预期目标
- 完成技术栈选型
- 实现生产级性能优化
- 完善监控和链路追踪

### Phase 3 预期目标  
- 插件系统落地
- 企业级插件生态
- 完善开发工具链

---

**变更负责人**: AI-PRD编辑器  
**变更时间**: 2026-03-02  
**变更类型**: 重构 (Refactor)