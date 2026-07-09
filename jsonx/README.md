# jsonx

`jsonx` 提供 JSON 字符串编解码和 Pretty JSON 能力。

## 导入

```go
import "github.com/jackman0925/go-foundation/jsonx"
```

## 基础用法

```go
text, err := jsonx.MarshalToString(data)
err = jsonx.UnmarshalFromString(text, &target)
pretty, err := jsonx.Pretty(data)
```

## 注意事项

- `MustToString` 失败时会 panic，仅适合测试或初始化阶段；
- 业务路径建议使用返回 error 的方法。
