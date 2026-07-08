# geox

`geox` 提供地理位置计算工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/geox"
```

## 基础用法

```go
meters := geox.DistanceMeters(116.397, 39.908, 121.473, 31.230)
```

## 注意事项

- `DistanceMeters` 使用近似公式，适合常规距离估算；
- 高精度测绘场景应使用专业地理库。
