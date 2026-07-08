# response

`response` 定义标准 API 响应结构和分页响应结构。

## 导入

```go
import "github.com/jackman0925/go-foundation/response"
```

## 基础用法

```go
ok := response.NewOK(data)
fail := response.NewFail(err)
page := response.NewPage(list, total, pageNo, pageSize)
```

## 注意事项

- 第一版不绑定 Gin 或其他 Web 框架；
- Web 框架适配建议放在业务项目或单独适配层。
