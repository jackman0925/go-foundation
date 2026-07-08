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
