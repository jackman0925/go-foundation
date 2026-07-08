# Changelog

## v0.1.0

- 初始化 `go-foundation` module。
- 新增 `config`、`errors`、`response`、`pagination`、`timex`、`stringx`、`jsonx`、`crypto`、`idgen` 包。
- 从 `well-wishes-service/internal/utils` 归纳并优化通用能力：default tag、命名转换、密码哈希、map checksum、Snowflake ID 生成器。
- 新增各包单元测试和基础示例。
- 明确第一版不包含日志封装、数据库连接和事务、Redis 基础封装、Gin 通用中间件。
