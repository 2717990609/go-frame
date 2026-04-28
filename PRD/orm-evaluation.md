# ORM 框架选型评估

## 1. 背景

Phase 2 技术栈升级目标：评估 Gorm → SQLx + QueryBuilder 的迁移可行性，并建立性能基准。

## 2. 候选方案对比

| 维度 | Gorm (当前) | SQLx | database/sql + 自建 Builder |
|------|-------------|------|-----------------------------|
| **性能** | 中等，反射开销 | 高，手写 SQL/命名参数 | 最高，零抽象 |
| **类型安全** | 弱，反射 | 强，结构体 Scan | 强 |
| **学习曲线** | 低 | 中 | 高 |
| **生态** | 完善，迁移/关联 | 轻量 | 需自建 |
| **SQL 可控性** | 中，链式 API | 高，原生 SQL | 完全可控 |
| **迁移支持** | AutoMigrate | 需第三方 | 需自建 |
| **维护活跃度** | 高 | 中 | - |

## 3. 当前 Gorm 使用范围

```
- pkg/database/mysql.go      Provider 实现（内部 Gorm）
- pkg/framework/database.go  直接返回 *gorm.DB
- plugins/mysql/             插件创建 gorm.DB
- internal/repository/       业务层直接使用 gorm.DB
- internal/service/health_service.go  db.Exec("SELECT 1")
```

**关键依赖**：Repository 层、Service 层均依赖 `*gorm.DB`。

## 4. 迁移策略

### 4.1 渐进式迁移（推荐）

1. **保持 database.Provider 抽象**：现有 `database.Provider` 已封装 CRUD，可支撑多实现。
2. **新增 SQLx Provider**：实现 `database.Provider` 的 SQLx 版本，通过配置切换。
3. **Repository 层适配**：引入 `database.Provider` 接口，逐步替换 `*gorm.DB` 直接依赖。
4. **framework 兼容层**：`GetDB()` 可返回接口或保持 `*gorm.DB` 直至迁移完成。

### 4.2 迁移优先级

| 优先级 | 任务 | 工作量 |
|--------|------|--------|
| P0 | SQLx Provider 实现 Provider 接口 | 中 |
| P1 | Repository 改为使用 Provider | 低 |
| P2 | 迁移工具：Gorm 模型 → SQLx 结构体 | 高 |
| P3 | 移除 Gorm 依赖 | 低 |

## 5. 性能考量与基准结果

**基准测试**：`go test -bench='Benchmark(Gorm|SQLx|Database)' -benchmem ./tests/benchmark/`

| 操作 | Gorm | SQLx | database/sql |
|------|------|------|--------------|
| Insert | ~11µs, 6444B, 94 allocs | ~2µs, 264B, 11 allocs | - |
| Find/Query | ~4.8µs, 4653B, 69 allocs | ~2.6µs, 856B, 29 allocs | ~2.1µs, 560B, 21 allocs |

**结论**：SQLx 在 Insert 上约 5.5x 更快、内存约 24x 更低；Find 约 1.8x 更快、内存约 5x 更低。

## 6. 结论与建议

- **短期**：保留 Gorm，完善 `database.Provider` 抽象与测试，为后续切换做准备。
- **中期**：实现 SQLx Provider，支持配置切换，在性能敏感路径优先使用。
- **长期**：业务层全面迁移至 Provider 接口后，可按需移除 Gorm。

## 7. 参考

- [Gorm 文档](https://gorm.io/docs/)
- [SQLx 文档](https://jmoiron.github.io/sqlx/)
- PRD: `froad-refactor.md` Phase 2
