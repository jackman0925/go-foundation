# netx

`netx` 提供 URL、HTTP 和本机网络接口相关的小工具。

## 导入

```go
import "github.com/jackman0925/go-foundation/netx"
```

## 基础用法

```go
domain, err := netx.Domain("https://example.com:8443/a/b?x=1")
joined, err := netx.URLPathJoin("https://example.com/api/", "/v1/", "users?active=true")
clientIP := netx.ClientIPFromHTTPRequest(request)
err := netx.ShutdownHTTPServer(shutdownCtx, server)
interfaces, err := netx.LocalIPv4Interfaces(netx.LocalInterfaceOptions{})
ips := netx.LocalIPsFromInterfaces(interfaces, bindAddr, showAll)
subnets := netx.InterfacesBySubnet(interfaces, bindAddr, showAll)
```

## 注意事项

### 收集 LAN 地址候选

Windows 的 `vEthernet` 可能承载实际可用的局域网地址。需要展示完整候选时使用已有选项：

```go
interfaces, err := netx.LocalIPv4Interfaces(netx.LocalInterfaceOptions{
	IncludeVirtual: true,
})
if err != nil {
	return err
}
for _, iface := range interfaces {
	if netx.IsVirtualInterfaceName(iface.Name) {
		// TODO: 展示为其他候选地址，提示需要确认；不要直接丢弃。
		continue
	}
	// TODO: 展示普通候选地址，同样需要客户端验证连通性。
}
```

此选项不会放开回环、未启用接口和 RFC2544 地址的默认过滤。应始终允许查看其他候选，不能仅在普通候选为空时才收集。

名称匹配与私有 IP 地址段都不能证明可达性；实际访问还取决于服务监听地址、防火墙和网络路径。`localhost` 回退只是展示行为，不表示机器没有 LAN 地址。

### 其他约定

- `Domain` 返回 `scheme://host[:port]`；
- `URLPathJoin` 保留第一个非空 scheme 和 host，并使用最后一个非空 query；
- `ClientIPFromHTTPRequest` 只解析 `RemoteAddr`，不信任代理头，避免在公共库中隐式接受可伪造来源。
- `ShutdownHTTPServer` 会先等待在途 HTTP 请求完成；上下文超时或取消后会关闭活动连接。被 Hijack 的连接（例如 WebSocket）需要业务项目自行关闭。
- `LocalIPv4Interfaces` 默认跳过 loopback、inactive、常见虚拟/隧道接口和 RFC2544 benchmark 地址。
