# pathx

跨平台词法路径处理。错误可通过 errors.Is 与 ErrInvalidPath、ErrPathEscape 比较。

```go
urlPath, err := pathx.NormalizeURLPath("/资料/%E6%96%87%E4%BB%B6.txt")
if err != nil {
	return err
}
// URL 只在入口解码一次。文件系统接口接收已解码的相对路径。
fullPath, err := pathx.ResolveUnderRoot(root, strings.TrimPrefix(urlPath, "/"))
if err != nil {
	return err
}
// TODO: 在执行文件访问前检查业务授权，并落实符号链接与并发目录替换防护。
_ = fullPath
```

上传相对路径使用 SanitizeRelativePath；它不会把 % 编码再次解码。
ResolveUnderRoot 的空相对路径或 . 表示根目录；SanitizeRelativePath 则拒绝空路径和仅 .。

本包不解析符号链接，不应将词法包含关系视为安全打开文件的保证。
统一拒绝 Windows 保留名和危险路径形式，因此部分 Unix 合法文件名也会被拒绝。
迁移时需注意：非法百分号编码、任意 .. 段和 Windows 危险名称均会被拒绝，
与原先容错解码或自动折叠 .. 的实现不兼容。
NormalizeURLPath 接收编码后的 URL 路径而非完整 URL；例如 %252e%252e 仅解码为字面量 %2e%2e。
ResolveUnderRoot 要求 root 非空，返回绝对路径。业务授权、内部目录限制及符号链接防护由调用方负责。
