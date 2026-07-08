package main

import (
	"os"
	"path/filepath"

	"github.com/jackman0925/go-foundation/config"
)

type appConfig struct {
	App struct {
		Name string `yaml:"name" default:"demo"`
		Env  string `yaml:"env" required:"true"`
	} `yaml:"app"`
}

func main() {
	path := filepath.Join(os.TempDir(), "go-foundation-config-example.yaml")
	if err := os.WriteFile(path, []byte("app:\n  env: dev\n"), 0o600); err != nil {
		panic(err)
	}

	cfg := config.MustLoad[appConfig](path)
	_ = cfg
}
