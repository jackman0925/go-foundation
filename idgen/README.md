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

## 注意事项

- `MachineID` 必须在 `0` 到 `1023` 之间；
- 生成器使用实例状态，不使用全局机器号，避免不同项目隐式互相影响；
- 如果系统时钟回拨，`NextID` 会返回错误，调用方应显式处理。
