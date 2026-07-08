# config

`config` 负责加载 YAML 配置，并支持默认值、环境变量覆盖和必填校验。

## 导入

```go
import "github.com/jackman0925/go-foundation/config"
```

## 基础用法

```go
type AppConfig struct {
    App struct {
        Name string `yaml:"name" env:"APP_NAME" default:"demo"`
        Port int    `yaml:"port" env:"APP_PORT" default:"8080"`
        Env  string `yaml:"env" required:"true"`
    } `yaml:"app"`
}

cfg := config.MustLoad[AppConfig]("configs/config.yaml")

type Request struct {
    Name *string `default:"guest"`
}

req := &Request{}
err := config.ApplyDefaults(req)
```

## 注意事项

- 当前支持基础标量类型：string、int、uint、float、bool；
- `ApplyDefaults` 支持基础类型指针和嵌套结构体；
- `env` 会覆盖 YAML 和 default；
- `required:"true"` 在默认值和环境变量处理后校验。
