# pagination

`pagination` 负责解析分页参数并生成 `limit`、`offset`。

## 导入

```go
import "github.com/jackman0925/go-foundation/pagination"
```

## 基础用法

```go
page := pagination.Parse("1", "20")
limit, offset := page.LimitOffset()
totalPages := pagination.TotalPages(41, 20)
```

## 注意事项

- 默认页码为 `1`；
- 默认页大小为 `20`；
- 最大页大小为 `100`，用于避免无界查询；
- `TotalPages` 在 `totalRecords <= 0` 或 `pageSize <= 0` 时返回 `0`。
