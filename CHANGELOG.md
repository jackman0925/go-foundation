# Changelog

## v0.1.0

- 初始化 `go-foundation` module。
- 新增 `config`、`errors`、`response`、`pagination`、`timex`、`stringx`、`jsonx`、`crypto`、`idgen` 包。
- 从 `well-wishes-service/internal/utils` 归纳并优化通用能力：default tag、命名转换、密码哈希、map checksum、Snowflake ID 生成器。
- 从 `pay_assist/pkg/utils` 归纳并优化通用能力：16 位 MD5、总页数计算、字符串校验、版本比较、URL 处理、HTTP 客户端 IP 提取。
- 从 `pay_assist/pkg/crypto/aes.go` 归纳 AES 能力，并改良为显式 key 的 AES-GCM 加解密。
- 新增各包单元测试和基础示例。
