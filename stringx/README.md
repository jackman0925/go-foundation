# stringx

`stringx` 提供字符串判空、截断、脱敏、随机字符串、命名转换、校验和版本比较能力。

## 导入

```go
import "github.com/jackman0925/go-foundation/stringx"
```

## 基础用法

```go
blank := stringx.IsBlank(" ")
masked := stringx.MaskMobile("13800138000")
short := stringx.Truncate("你好世界", 2)
token, err := stringx.RandomString(16)
field := stringx.CamelToSnake("UserID")
emailOK := stringx.IsEmail("user@example.com")
numeric := stringx.IsNumeric("123")
joined := stringx.JoinInt64([]int64{1, 2, 3}, ",")
cmp := stringx.CompareVersion("1.2.0", "1.1.9")
```

## 注意事项

- `Truncate` 按 rune 截断，不会切断中文字符；
- `CamelToSnake` 会处理 `UserID`、`HTTPServer`、`APIKey` 等连续大写场景；
- `CompareVersion` 支持 `.` 和 `_` 分隔的纯数字版本段；
- `RandomString` 使用 `crypto/rand`。
