package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Load 读取 YAML 文件到 T，并应用 default、env 和 required 标签。
func Load[T any](path string) (*T, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg T
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := ApplyDefaults(&cfg); err != nil {
		return nil, err
	}
	if err := applyEnvAndRequired(reflect.ValueOf(&cfg).Elem(), ""); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MustLoad 读取 YAML 文件到 T，加载失败时 panic。
func MustLoad[T any](path string) *T {
	cfg, err := Load[T](path)
	if err != nil {
		panic(err)
	}
	return cfg
}

// ApplyDefaults 将 default 标签应用到结构体指针中的零值字段。
func ApplyDefaults(target any) error {
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}

	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return applyDefaults(value, "")
}

func applyDefaults(value reflect.Value, path string) error {
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
		if field.Kind() == reflect.Pointer && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := applyDefaults(field.Elem(), fieldPath); err != nil {
				return err
			}
			continue
		}
		if field.Kind() == reflect.Struct {
			if err := applyDefaults(field, fieldPath); err != nil {
				return err
			}
			continue
		}

		if defaultValue, ok := fieldType.Tag.Lookup("default"); ok && defaultValue != "-" {
			if field.Kind() == reflect.Pointer {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
					if err := setFromString(field.Elem(), defaultValue); err != nil {
						return fmt.Errorf("apply default for %s: %w", fieldPath, err)
					}
				}
			} else if isZero(field) {
				if err := setFromString(field, defaultValue); err != nil {
					return fmt.Errorf("apply default for %s: %w", fieldPath, err)
				}
			}
		}
	}

	return nil
}

func applyEnvAndRequired(value reflect.Value, path string) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
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
		if field.Kind() == reflect.Pointer && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := applyEnvAndRequired(field.Elem(), fieldPath); err != nil {
				return err
			}
			continue
		}
		if field.Kind() == reflect.Struct {
			if err := applyEnvAndRequired(field, fieldPath); err != nil {
				return err
			}
			continue
		}

		if envName := fieldType.Tag.Get("env"); envName != "" {
			if envValue, ok := os.LookupEnv(envName); ok {
				target := field
				if target.Kind() == reflect.Pointer {
					if target.IsNil() {
						target.Set(reflect.New(target.Type().Elem()))
					}
					target = target.Elem()
				}
				if err := setFromString(target, envValue); err != nil {
					return fmt.Errorf("apply env %s for %s: %w", envName, fieldPath, err)
				}
			}
		}

		if fieldType.Tag.Get("required") == "true" && isEmptyConfigValue(field) {
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
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(parsed)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}

	return nil
}

func isZero(value reflect.Value) bool {
	return value.IsZero()
}

func isEmptyConfigValue(value reflect.Value) bool {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return true
		}
		return isEmptyConfigValue(value.Elem())
	}
	return value.IsZero()
}

func joinPath(prefix string, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
