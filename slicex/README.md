# slicex

`slicex` 提供泛型 slice 小工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/slicex"
```

## 基础用法

```go
ok := slicex.Contains([]int{1, 2, 3}, 2)
unique := slicex.Unique([]int{3, 1, 3, 2})
slicex.Reverse(unique)
```

## 注意事项

- `Unique` 保留首次出现顺序；
- `Reverse` 原地修改传入切片。
