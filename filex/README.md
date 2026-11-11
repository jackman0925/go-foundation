# filex

`filex` 提供低风险、标准库实现的文件操作工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/filex"
```

## 基础用法

```go
ok := filex.Exists("config.yaml")
isFile := filex.IsFile("config.yaml")
isDir := filex.IsDir("configs")

err := filex.EnsureDir("tmp/output")
err = filex.WriteText("tmp/output/demo.txt", "hello", 0o600)
text, err := filex.ReadText("tmp/output/demo.txt")
err = filex.CopyFile("tmp/output/demo.txt", "tmp/output/copy.txt")

size, err := filex.FileSize("tmp/output/demo.txt")
display := filex.FormatDecimalSize(size)
diagnostic := filex.FormatBinarySize(size)
ext := filex.Ext("archive.tar.gz")
name := filex.BaseName("/tmp/archive.tar.gz")
```

## 注意事项

大小展示的单位、舍入及跨语言样例见 [文件大小展示规范](FILE_SIZE_DISPLAY.md)。展示文本不用于上传限制和配额计算。

- `WriteText` 和 `CopyFile` 会自动创建目标父目录；
- `CopyFile` 只复制普通文件，不复制目录或符号链接语义；
- 本包不提供递归删除、下载、文件锁等高风险或重能力。
