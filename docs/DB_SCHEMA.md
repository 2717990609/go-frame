# 数据库表结构

> 使用 migrate 工具执行 migrations/ 下的脚本，禁止手动改表。

## 表清单

| 表名 | 说明 |
|------|------|
| sd_user | 用户表（示例） |

## 迁移命令

```bash
# 使用 golang-migrate 示例
migrate -path migrations -database "mysql://user:pass@tcp(host:3306)/db" up
```
