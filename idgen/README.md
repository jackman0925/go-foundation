# idgen

`idgen` 提供可配置的 Snowflake ID 生成器。

## 导入

```go
import "github.com/jackman0925/go-foundation/idgen"
```

## 基础用法

```go
generator, err := idgen.NewSnowflake(idgen.SnowflakeOptions{
    MachineID: 1,
})
if err != nil {
    return err
}

id, err := generator.NextID()
```

## 老项目适配示例

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
- 生成器使用实例状态，不使用全局机器号，避免不同项目隐式互相影响；
- 如果系统时钟回拨，`NextID` 会返回错误，调用方应显式处理。
