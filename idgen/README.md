# idgen

`idgen` 提供可配置的 Snowflake ID 生成器，以及适合低并发场景的前端安全数字 ID 生成器。

## 导入

```go
import "github.com/jackman0925/go-foundation/idgen"
```

## Snowflake 基础用法

```go
generator, err := idgen.NewSnowflake(idgen.SnowflakeOptions{
    MachineID: 1,
})
if err != nil {
    return err
}

id, err := generator.NextID()
```

## SafeNumberID 基础用法

`SafeNumberID` 生成纯数字 ID，默认 14 位，数值小于 JavaScript `Number.MAX_SAFE_INTEGER`，前端用 `number` 接收也不会丢失精度。

```go
generator, err := idgen.NewSafeNumberID(idgen.SafeNumberIDOptions{})
if err != nil {
    return err
}

id, err := generator.Next()
text, err := generator.NextString()
```

默认格式为：

```text
YYDDDHHMMSS + SSS
```

示例：`26224090807000` 表示 2026 年第 224 天 09:08:07，同秒内序列为 `000`。

可按低并发容量调整同秒序列位数：

```go
generator, err := idgen.NewSafeNumberID(idgen.SafeNumberIDOptions{
    SequenceDigits: 4,
})
```

`SequenceDigits` 可设置为 `1` 到 `4`。默认 `3` 表示同一实例同一秒最多生成 `1000` 个 ID；最大 `4` 表示最多 `10000` 个，仍保持 JS 安全整数范围内。

## 老项目 Snowflake 适配示例

老项目如果原来习惯直接调用 `utils.GenerateID()`，建议在业务项目中新增一个很薄的适配包。全局实例放在业务项目里，`go-foundation` 不内置全局生成器，避免隐藏 `MachineID` 配置。

```go
package id

import (
    "errors"

    "github.com/jackman0925/go-foundation/idgen"
)

var generator *idgen.Snowflake

func Init(machineID int64) error {
    g, err := idgen.NewSnowflake(idgen.SnowflakeOptions{
        MachineID: machineID,
    })
    if err != nil {
        return err
    }
    generator = g
    return nil
}

func NextID() (int64, error) {
    if generator == nil {
        return 0, errors.New("id generator is not initialized")
    }
    return generator.NextID()
}
```

启动阶段从配置传入 `MachineID`：

```go
if err := id.Init(cfg.App.MachineID); err != nil {
    return err
}
```

业务代码中调用：

```go
orderID, err := id.NextID()
```

## 注意事项

- `MachineID` 必须在 `0` 到 `1023` 之间；
- Snowflake 生成器使用实例状态，不使用全局机器号，避免不同项目隐式互相影响；
- 如果系统时钟回拨，Snowflake 的 `NextID` 会返回错误，调用方应显式处理；
- `SafeNumberID` 只保证同一生成器实例内不会重复，不保证跨进程、跨机器全局唯一；
- `SafeNumberID` 同一秒序列耗尽时会返回错误，不会隐藏等待；
- 核心业务 ID、跨服务唯一 ID 仍建议使用 Snowflake，并在面向前端的 API 中按字符串返回。
