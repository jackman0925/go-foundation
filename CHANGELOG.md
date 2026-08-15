# Changelog

本文档记录 go-foundation Lib的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [0.1.3] - 2026-08-15

- 新增 `mailx` 包，提供基础 SMTP 邮件发送能力。
- 支持 `To`、`Cc`、`Bcc` 多收件人场景，`Bcc` 只参与 SMTP 投递，不写入 MIME 邮件头。
- 支持纯文本、HTML、文本加 HTML 混合正文和附件发送。
- 新增 `mailx.Sender` 接口，便于业务项目替换实现或单元测试 mock。
- 新增 `mailx.LimitedSender`，支持对邮件发送进行实例级并发限制和等待队列限制。
- 新增 `mailx.ErrMailBusy`，当并发和等待队列都满时明确返回错误，邮件不会静默丢弃。
- 新增 `mailx.NewSMTPClient`，提供默认带限流保护的 SMTP 推荐入口。
- SMTP 发送支持超时、最大邮件大小限制、明文、STARTTLS 和隐式 TLS 连接方式。
- 补充 `mailx` 单元测试和本地假 SMTP 服务测试，覆盖 MIME 构建、收件人去重、密送保护、并发限制、等待队列满和 SMTP 投递流程。

## [0.1.2] - 2026-08-12

- 新增 `idgen.SafeNumberID`，提供前端 JS 安全的本地数字 ID 生成器。
- 支持 `SafeNumberID.Next()` 返回 `int64`，以及 `SafeNumberID.NextString()` 返回数字字符串。
- `SafeNumberID` 支持可配置同秒序列位数，默认 3 位，最大 4 位，保持在 JavaScript 安全整数范围内。
- 补充 `SafeNumberID` 单元测试，覆盖格式、并发唯一性、序列溢出和 JS 安全整数边界。
- 更新 `idgen` 文档，明确 `SafeNumberID` 只保证同一生成器实例内不重复，不替代跨机器全局唯一 ID。

## [0.1.1] - 2026-07-10

- 新增 `filex`，提供文件存在性、文本读写、文件复制、大小和路径信息工具。
- 优化 `netx` 本机网卡筛选，补充 Windows 下 VMware、VirtualBox、Hyper-V、vEthernet 等虚拟网卡名称识别。

## [0.1.0] - 2026-07-08

- 初始化 `go-foundation` module。
- 新增 `config`、`errors`、`response`、`pagination`、`timex`、`stringx`、`jsonx`、`crypto`、`idgen` 包。
- default tag、命名转换、密码哈希、map checksum、Snowflake ID 生成器。
- 16 位 MD5、总页数计算、字符串校验、版本比较、URL 处理、HTTP 客户端 IP 提取、本机 IPv4 网卡筛选。
- AES 加解密、gzip/zlib、CRC16、slice helpers、颜色转换、经纬度距离计算。
- 新增各包单元测试和基础示例。
