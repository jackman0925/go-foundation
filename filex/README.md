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

## 同目录提交写入

```go
err := filex.WriteAtomic("config.json", data, filex.AtomicWriteOptions{
	Mode: 0o600,
})
written, err := filex.CopyAtomic("uploads/file.bin", reader, filex.AtomicWriteOptions{
	CreateDir: true,
})
```

默认 Mode 为 0644，执行文件 Sync，不自动创建父目录。CopyAtomic 不关闭 reader；
错误时 written 只表示复制到临时文件的字节数，不代表提交成功。不会先删除旧目标。
非 Unix 平台不保证原子重命名，也不承诺断电持久性或防御符号链接竞态。
Mode 仅接受普通权限位；CreateDir 使用 0755 创建父目录（受 umask 影响）。
提交前失败会尽力清理临时文件，清理错误不会覆盖原始错误；并发写入以最后成功提交者为准。
调用方必须信任目标父目录，本接口不负责授权、版本保留或冲突检测。

## 注意事项

大小展示的单位、舍入及跨语言样例见 [文件大小展示规范](FILE_SIZE_DISPLAY.md)。展示文本不用于上传限制和配额计算。

- `WriteText` 和 `CopyFile` 会自动创建目标父目录；
- `CopyFile` 只复制普通文件，不复制目录或符号链接语义；
- 本包不提供递归删除、下载、文件锁等高风险或重能力。
