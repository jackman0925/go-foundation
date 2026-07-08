# timex

`timex` 提供常用时间格式化、解析和日/月边界计算。

## 导入

```go
import "github.com/jackman0925/go-foundation/timex"
```

## 基础用法

```go
start := timex.StartOfDay(time.Now())
end := timex.EndOfDay(time.Now())
text := timex.FormatDateTime(time.Now())
```

## 注意事项

- 日/月边界会保留输入时间的 location；
- `Parse` 支持常见日期、日期时间和 RFC3339 格式。
