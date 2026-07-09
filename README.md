# go-foundation

公司级 Go 基础库，沉淀可公开复用的通用工具能力，避免在项目之间复制 `utils`。

## 安装

```bash
go get github.com/jackman0925/go-foundation@v0.1.0
```

## 包列表

| 包 | 作用 |
| --- | --- |
| `config` | YAML 配置加载、默认值、环境变量覆盖、必填校验 |
| `errors` | 统一错误码、错误消息、错误包装 |
| `response` | 标准 API 响应结构和分页响应结构 |
| `pagination` | 分页参数解析、limit/offset 计算、总页数计算 |
| `timex` | 日期时间格式化、解析、日/月边界 |
| `stringx` | 字符串判空、截断、脱敏、随机字符串、命名转换、校验和版本比较 |
| `jsonx` | JSON 字符串编解码、Pretty JSON |
| `crypto` | MD5、SHA256、HMAC-SHA256、随机数字、密码哈希、map checksum |
| `idgen` | 可配置 Snowflake ID 生成器 |
| `netx` | URL 域名提取、URL path 拼接、HTTP 客户端 IP 提取、本机 IPv4 网卡筛选 |
| `compressx` | gzip、zlib 压缩和解压 |
| `slicex` | 泛型 slice 包含、去重、反转 |
| `colorx` | 十六进制颜色和 RGBA 转换 |
| `geox` | 经纬度距离计算 |

## 快速开始

```go
cfg := config.MustLoad[AppConfig]("configs/config.yaml")

err := foundationErrors.New(foundationErrors.CodeBadRequest, "参数错误")
resp := response.NewFail(err)

page := pagination.Parse("1", "20")
limit, offset := page.LimitOffset()

_ = cfg
_ = resp
_ = limit
_ = offset
```

## 不适合放入本库的内容

- 具体业务代码；
- 公司内部接口地址、密钥、Token、证书、账号；
- 真实业务表结构、客户名称、内部权限模型；
- 支付、License、微信、管理后台等业务封装。

## 第一版不包含

`v0.1.0` 不做日志封装、数据库连接和事务、Redis 基础封装、Gin 通用中间件。

本库参考业务项目 utils 时，只归纳通用且可测试的能力。


## 开发规范

- Go 版本：`1.23.12`；
- 所有非 trivial 逻辑必须有单元测试；
- 导出类型和导出函数必须有注释；
- 禁止使用 `println`、`fmt.Println` 代替项目日志；
- 执行 `gofmt` 和 `go test ./...` 后再提交。

## License

见 [LICENSE](LICENSE)。
