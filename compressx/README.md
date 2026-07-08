# compressx

`compressx` 提供 gzip 和 zlib 压缩/解压能力。

## 导入

```go
import "github.com/jackman0925/go-foundation/compressx"
```

## 基础用法

```go
compressed, err := compressx.Gzip([]byte("hello"))
plain, err := compressx.Gunzip(compressed)

zdata, err := compressx.Zlib([]byte("hello"))
plain, err = compressx.Unzlib(zdata)
```

## 注意事项

- 解压非法数据会返回 error；
- 函数不吞错误，调用方必须显式处理。
