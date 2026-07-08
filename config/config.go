package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Load reads a YAML file into T, then applies default, env, and required tags.
func Load[T any](path string) (*T, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg T
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := applyTags(reflect.ValueOf(&cfg).Elem(), ""); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MustLoad reads a YAML file into T and panics if loading fails.
func MustLoad[T any](path string) *T {
	cfg, err := Load[T](path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func applyTags(value reflect.Value, path string) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}

		fieldPath := joinPath(path, fieldType.Name)
		if field.Kind() == reflect.Struct {
			if err := applyTags(field, fieldPath); err != nil {
				return err
			}
			continue
		}

		if isZero(field) {
			if defaultValue, ok := fieldType.Tag.Lookup("default"); ok {
				if err := setFromString(field, defaultValue); err != nil {
					return fmt.Errorf("apply default for %s: %w", fieldPath, err)
				}
			}
		}

		if envName := fieldType.Tag.Get("env"); envName != "" {
			if envValue, ok := os.LookupEnv(envName); ok {
				if err := setFromString(field, envValue); err != nil {
					return fmt.Errorf("apply env %s for %s: %w", envName, fieldPath, err)
				}
			}
		}

		if fieldType.Tag.Get("required") == "true" && isZero(field) {
			return fmt.Errorf("required config field %s is empty", fieldPath)
		}
	}

	return nil
}

func setFromString(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}

	return nil
}

func isZero(value reflect.Value) bool {
	return value.IsZero()
}

func joinPath(prefix string, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
