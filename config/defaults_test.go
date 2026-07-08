package config

import "testing"

func TestApplyDefaultsSupportsPointersAndNestedStructs(t *testing.T) {
	type nested struct {
		Enabled *bool `default:"true"`
	}
	type request struct {
		Name   string  `default:"guest"`
		Age    *int    `default:"18"`
		Score  float64 `default:"9.5"`
		Nested nested
	}

	var cfg request
	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults returned error: %v", err)
	}

	if cfg.Name != "guest" {
		t.Fatalf("expected default name, got %q", cfg.Name)
	}
	if cfg.Age == nil || *cfg.Age != 18 {
		t.Fatalf("expected default age pointer, got %v", cfg.Age)
	}
	if cfg.Score != 9.5 {
		t.Fatalf("expected default score, got %f", cfg.Score)
	}
	if cfg.Nested.Enabled == nil || *cfg.Nested.Enabled != true {
		t.Fatalf("expected nested default bool pointer, got %v", cfg.Nested.Enabled)
	}
}

func TestApplyDefaultsRejectsInvalidDefaultValue(t *testing.T) {
	type request struct {
		Age int `default:"bad"`
	}

	var cfg request
	if err := ApplyDefaults(&cfg); err == nil {
		t.Fatal("expected invalid default value error")
	}
}

func TestApplyDefaultsRejectsUnsupportedDefaultType(t *testing.T) {
	type request struct {
		Items []string `default:"a,b"`
	}

	var cfg request
	if err := ApplyDefaults(&cfg); err == nil {
		t.Fatal("expected unsupported default type error")
	}
}

func TestApplyDefaultsRejectsInvalidInput(t *testing.T) {
	if err := ApplyDefaults(nil); err == nil {
		t.Fatal("expected nil target error")
	}
	if err := ApplyDefaults(struct{}{}); err == nil {
		t.Fatal("expected non-pointer target error")
	}
	value := 1
	if err := ApplyDefaults(&value); err == nil {
		t.Fatal("expected pointer to non-struct error")
	}
}

func TestApplyDefaultsSupportsUnsignedAndBoolTypes(t *testing.T) {
	type request struct {
		Count uint `default:"10"`
		Flag  bool `default:"true"`
	}

	var cfg request
	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults returned error: %v", err)
	}

	if cfg.Count != 10 {
		t.Fatalf("expected count 10, got %d", cfg.Count)
	}
	if !cfg.Flag {
		t.Fatal("expected flag true")
	}
}

func TestApplyDefaultsSupportsNestedStructPointers(t *testing.T) {
	type nested struct {
		Name string `default:"inner"`
	}
	type request struct {
		Nested *nested
	}

	var cfg request
	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults returned error: %v", err)
	}

	if cfg.Nested == nil {
		t.Fatal("expected nested struct pointer to be initialized")
	}
	if cfg.Nested.Name != "inner" {
		t.Fatalf("expected nested default, got %q", cfg.Nested.Name)
	}
}

func TestLoadAppliesEnvInsideNestedStructPointer(t *testing.T) {
	type nested struct {
		Name string `env:"INNER_NAME"`
	}
	type request struct {
		Nested *nested
	}

	t.Setenv("INNER_NAME", "env-inner")
	path := writeTempConfig(t, "{}\n")

	cfg, err := Load[request](path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Nested == nil || cfg.Nested.Name != "env-inner" {
		t.Fatalf("expected nested pointer env value, got %+v", cfg.Nested)
	}
}

func TestLoadRejectsEmptyRequiredPointerValue(t *testing.T) {
	type request struct {
		Name *string `default:"" required:"true"`
	}

	path := writeTempConfig(t, "{}\n")
	_, err := Load[request](path)
	if err == nil {
		t.Fatal("expected required pointer value error")
	}
}

func TestLoadReturnsEnvConversionError(t *testing.T) {
	type request struct {
		Port int `env:"APP_PORT"`
	}

	t.Setenv("APP_PORT", "bad")
	path := writeTempConfig(t, "{}\n")
	_, err := Load[request](path)
	if err == nil {
		t.Fatal("expected env conversion error")
	}
}

func TestLoadAppliesEnvForUnsignedAndBoolFields(t *testing.T) {
	type request struct {
		Count uint `env:"APP_COUNT"`
		Flag  bool `env:"APP_FLAG"`
	}

	t.Setenv("APP_COUNT", "42")
	t.Setenv("APP_FLAG", "true")
	path := writeTempConfig(t, "{}\n")

	cfg, err := Load[request](path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Count != 42 {
		t.Fatalf("expected count 42, got %d", cfg.Count)
	}
	if !cfg.Flag {
		t.Fatal("expected flag true")
	}
}

func TestLoadRejectsMissingRequiredPointer(t *testing.T) {
	type request struct {
		Name *string `required:"true"`
	}

	path := writeTempConfig(t, "{}\n")
	if _, err := Load[request](path); err == nil {
		t.Fatal("expected missing required pointer error")
	}
}
