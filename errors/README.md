# errors

`errors` 提供统一错误码、错误消息和错误包装能力。

## 导入

```go
import foundationErrors "github.com/jackman0925/go-foundation/errors"
```

## 基础用法

```go
err := foundationErrors.New(foundationErrors.CodeBadRequest, "参数错误")
err = foundationErrors.Wrap(foundationErrors.CodeInternal, "查询失败", err)
```

## 注意事项

- 本包只放通用错误码；
- 具体业务错误码应在业务项目中定义；
- `Wrap` 会保留原始 error，支持标准库 `errors.Is` 和 `errors.As`。
