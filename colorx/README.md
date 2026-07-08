# colorx

`colorx` 提供十六进制颜色和 `color.RGBA` 转换。

## 导入

```go
import "github.com/jackman0925/go-foundation/colorx"
```

## 基础用法

```go
rgba, err := colorx.HexToRGBA("#ff7ff0")
hex := colorx.RGBAToHex(rgba)
```

## 注意事项

- `HexToRGBA` 支持 `#RRGGBB` 和 `RRGGBB`；
- 不支持三位短写，例如 `#fff`。
