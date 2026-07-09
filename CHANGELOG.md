# Changelog

本文档记录 go-foundation Lib的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [未发布]

- 新增 `filex`，提供文件存在性、文本读写、文件复制、大小和路径信息工具。

## [0.1.0] - 2026-07-08

- 初始化 `go-foundation` module。
- 新增 `config`、`errors`、`response`、`pagination`、`timex`、`stringx`、`jsonx`、`crypto`、`idgen` 包。
- default tag、命名转换、密码哈希、map checksum、Snowflake ID 生成器。
- 16 位 MD5、总页数计算、字符串校验、版本比较、URL 处理、HTTP 客户端 IP 提取、本机 IPv4 网卡筛选。
- AES 加解密、gzip/zlib、CRC16、slice helpers、颜色转换、经纬度距离计算。
- 新增各包单元测试和基础示例。
