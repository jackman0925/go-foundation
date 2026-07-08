# netx

`netx` 提供 URL 和 HTTP 网络相关的小工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/netx"
```

## 基础用法

```go
domain, err := netx.Domain("https://example.com:8443/a/b?x=1")
joined, err := netx.URLPathJoin("https://example.com/api/", "/v1/", "users?active=true")
clientIP := netx.ClientIPFromHTTPRequest(request)
```

## 注意事项

- `Domain` 返回 `scheme://host[:port]`；
- `URLPathJoin` 保留第一个非空 scheme 和 host，并使用最后一个非空 query；
- `ClientIPFromHTTPRequest` 只解析 `RemoteAddr`，不信任代理头，避免在公共库中隐式接受可伪造来源。
