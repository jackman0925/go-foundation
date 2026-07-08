package config

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	App struct {
		Name string `yaml:"name" env:"APP_NAME" default:"demo"`
		Port int    `yaml:"port" env:"APP_PORT" default:"8080"`
		Mode string `yaml:"mode" required:"true"`
	} `yaml:"app"`
}

func TestLoadAppliesYAMLDefaultsEnvAndRequiredValidation(t *testing.T) {
	t.Setenv("APP_PORT", "9090")

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("app:\n  mode: test\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load[testConfig](path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.App.Name != "demo" {
		t.Fatalf("expected default app name, got %q", cfg.App.Name)
	}
	if cfg.App.Port != 9090 {
		t.Fatalf("expected env port override, got %d", cfg.App.Port)
	}
	if cfg.App.Mode != "test" {
		t.Fatalf("expected yaml mode, got %q", cfg.App.Mode)
	}
}

func TestLoadRejectsMissingRequiredField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("app:\n  name: demo\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load[testConfig](path)
	if err == nil {
		t.Fatal("expected missing required field error")
	}
}

func TestLoadReturnsReadAndParseErrors(t *testing.T) {
	if _, err := Load[testConfig](filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected missing file error")
	}

	path := writeTempConfig(t, "app:\n  mode: [bad\n")
	if _, err := Load[testConfig](path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestMustLoadPanicsOnError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = MustLoad[testConfig](filepath.Join(t.TempDir(), "missing.yaml"))
}

func TestMustLoadReturnsConfigOnSuccess(t *testing.T) {
	path := writeTempConfig(t, "app:\n  mode: test\n")

	cfg := MustLoad[testConfig](path)
	if cfg.App.Mode != "test" {
		t.Fatalf("expected mode test, got %q", cfg.App.Mode)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
