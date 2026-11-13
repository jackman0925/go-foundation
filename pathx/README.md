# pathx

跨平台词法路径处理及已存在目标的真实路径检查。错误可通过 errors.Is 与 ErrInvalidPath、ErrPathEscape 比较。

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

ResolveUnderRoot 不解析符号链接，不应将词法包含关系视为安全打开文件的保证。
统一拒绝 Windows 保留名和危险路径形式，因此部分 Unix 合法文件名也会被拒绝。
迁移时需注意：非法百分号编码、任意 .. 段和 Windows 危险名称均会被拒绝，
与原先容错解码或自动折叠 .. 的实现不兼容。
NormalizeURLPath 接收编码后的 URL 路径而非完整 URL；例如 %252e%252e 仅解码为字面量 %2e%2e。
ResolveUnderRoot 要求 root 非空，返回绝对路径。业务授权、内部目录限制及符号链接防护由调用方负责。

## 已存在目标的符号链接检查

```go
realPath, err := pathx.ResolveRealUnderRoot(root, relativePath)
if err != nil {
	return err
}
// TODO: 执行业务授权。只有目录树不会被不可信调用方修改时，才可依赖此次检查结果。
_ = realPath
```

- 根目录和目标必须存在；根目录必须是目录。不能用于新建上传文件路径的预授权。
- 允许指向根内的链接及链接链；最终真实路径越界返回 ErrPathEscape。
- 根目录本身可以是链接，解析后的真实目录作为信任边界；调用方必须信任根目录配置。
- 断链、循环、权限及不存在错误直接返回；不会为缺失目标自动回退到词法校验。
- 返回绝对真实路径，但检查过程及之后均可能发生目录替换（TOCTOU）。
  此接口只检查当时解析得到的位置，不保证后续 Open、Write 或 Remove 仍受根目录约束。
  对可被攻击者并发修改的目录，需要基于受限目录句柄的文件操作能力，不能仅靠返回路径字符串实现隔离。
- Windows 符号链接测试可能因权限不足跳过；交叉编译不代表已验证 junction 或其他 reparse point 的运行行为。

## 受限根目录操作

`ResolveUnderRoot` 和 `ResolveRealUnderRoot` 都返回路径字符串，不能消除检查与使用之间的竞态。需要把文件访问本身限制在工作目录内时，使用 `Root`：

```go
root, err := pathx.OpenRoot(workDir)
if err != nil {
	return err
}
defer root.Close()

if err := root.MkdirAll("incoming/2026", 0o750); err != nil {
	return err
}
file, err := root.OpenFile("incoming/2026/report.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
if err != nil {
	return err
}
defer file.Close()
```

- `Root` 在打开时持有根目录句柄；后续 `OpenDir`、`Open`、`OpenFile`、`MkdirAll`、`Remove`、`Rename` 都相对此句柄执行，不接受绝对路径或 `..`。`OpenDir("")` 可安全枚举受限根目录本身。
- macOS/Linux 使用 `openat` 逐段打开目录并加 `O_NOFOLLOW`；Windows 使用 `NtCreateFile` 的 `RootDirectory` 与 `OBJ_DONT_REPARSE`，拒绝符号链接、junction 和其他 reparse point。
- 根目录本身也不得为链接或 reparse point。根路径在对象创建后被替换时，已打开的根句柄仍指向原目录。
- `Remove` 不递归删除；`Rename` 仅允许同一个 `Root` 内移动，平台会替换同名目标。
- `OpenFile` 仅支持 `O_WRONLY`、`O_RDWR`、`O_CREATE`、`O_EXCL`、`O_TRUNC`；父目录必须先由 `MkdirAll` 创建。
- 目录句柄能力不处理业务 ACL、租户隔离、审计、文件冲突策略或递归删除。

Windows 的 reparse-point 运行测试需要在 Windows CI 执行；本机交叉编译只验证接口和构建，不替代真实系统测试。
