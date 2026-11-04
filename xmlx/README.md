# xmlx

基于 Go 标准库 `encoding/xml`，提供 XML 字符串编解码和缩进输出，无额外依赖。

## 使用

```go
import (
	"encoding/xml"

	"github.com/jackman0925/go-foundation/xmlx"
)

type User struct {
	XMLName xml.Name `xml:"user"`
	UserID int64 `xml:"userId,attr"`
	UserName string `xml:"userName"`
	Tags []string `xml:"tags>tag"`
}

func roundTrip() error {
	user := User{UserID: 1001, UserName: "张三", Tags: []string{"开发"}}
	text, err := xmlx.MarshalToString(user)
	if err != nil {
		return err
	}
	// 输出：<user userId="1001"><userName>张三</userName><tags><tag>开发</tag></tags></user>
	var result User
	if err := xmlx.UnmarshalFromString(text, &result); err != nil {
		return err
	}
	// TODO: 使用解码后的 result 执行业务逻辑。
	return nil
}
```

## 行为边界

- 字段需导出；使用 `xml` 标签指定节点、属性、嵌套路径或字符数据，`json` 标签不生效。
- `Pretty(value)` 编码结构体等值并使用两空格缩进，不是原始 XML 文本格式化工具。
- 不自动添加 XML 声明；确有需要时调用方可以拼接 `xml.Header`。
- XML 与 JSON 不同，标准库不支持直接编码 map；列表建议放在带根节点的结构体中。
- 解码需传入非空指针；未知字段等行为遵循标准库，不提供 XSD 校验或完整文档验证。
- 所有方法返回错误，不提供会主动 panic 的 Must 方法。
- 输入大小由调用方限制；大文件和流式解析直接使用 `xml.Decoder`。
