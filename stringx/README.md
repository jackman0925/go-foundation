# stringx

`stringx` 提供字符串判空、截断、脱敏、随机字符串和命名转换能力。

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
```

## 注意事项

- `Truncate` 按 rune 截断，不会切断中文字符；
- `CamelToSnake` 会处理 `UserID`、`HTTPServer`、`APIKey` 等连续大写场景；
- `RandomString` 使用 `crypto/rand`。
