package env

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"gopkg.in/yaml.v2"
)

// Config holds the configuration parameters for retrieving a config.
type Env struct {
	Name         string      // Name of the config file
	Type         ConfigType  // Type of the config file ("json" or "yaml")
	Path         string      // Path to the config file
	EnvPrefix    string      // Prefix for environment variable names
	ConfigStruct interface{} // Struct to store the config values
	SkipFields   []string    // Fields to skip when loading from env
}

// EnvType represents the environment type.
type EnvType string

const (
	LOCAL EnvType = "local" // Local environment will use the keychain
	DEV   EnvType = "dev"   // Dev environment will use the keychain
	PROD  EnvType = "prod"  // Prod environment will use the config file or env vars
)

type ConfigType string

const (
	JSON ConfigType = "json"
	YAML ConfigType = "yaml"
)

var CUSTOM ConfigType = "" // This will search for whatever custom tag you specify

// Get retrieves the configuration based on the specified environment type.
func Get(t EnvType, c *Env) error {
	switch t {
	case DEV, LOCAL:
		return BuildDevEnv(c, nil)
	case PROD:
		return LoadConfig(c)
	default:
		return fmt.Errorf("unsupported environment type: %s", t)
	}
}

// LoadConfig loads the config from the file or falls back to environmental variables.
func LoadConfig(c *Env) error {
	file, err := os.Open(c.Name)
	if err != nil {
		return c.loadFromEnv() // If the file doesn't exist, fallback to environmental variables
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	switch strings.ToLower(string(c.Type)) {
	case string(JSON):
		err = json.Unmarshal(data, &c.ConfigStruct)
	case string(YAML):
		err = yaml.Unmarshal(data, c.ConfigStruct)
	default:
		return fmt.Errorf("unsupported file type: %s", c.Type)
	}

	return err
}

func (c *Env) loadFromEnv() error {
	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()
	for i := 0; i < valueOf.NumField(); i++ {
		fieldType := typeOf.Field(i)

		if helpers.Contains(c.SkipFields, fieldType.Name) {
			i-- // Decrement the counter so we don't skip the next field
			continue
		}

		field := valueOf.Field(i)

		var envName string
		if tag := fieldType.Tag.Get(string(c.Type)); tag != "" {
			envName = strings.ToUpper(tag)
		} else if tag := fieldType.Tag.Get("yaml"); tag != "" {
			envName = strings.ToUpper(tag)
		} else {
			envName = strings.ToUpper(fieldType.Tag.Get("env"))
			if envName == "" {
				envName = fieldType.Name
			}
		}

		if c.EnvPrefix != "" {
			envName = c.EnvPrefix + "_" + envName
		}

		if envValue := os.Getenv(envName); envValue != "" {
			err := setField(field, envValue)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %s", fieldType.Name, err)
			}
		}
	}

	return nil
}

// setField sets the value of a field in the struct based on its type.
func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse integer value: %s", err)
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse unsigned integer value: %s", err)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse boolean value: %s", err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value: %s", err)
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported field type")
	}

	return nil
}

// GetKeys returns the keys for the struct.
func (c *Env) GetKeys() []string {
	var keys []string

	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()

	for i := 0; i < valueOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		if !helpers.Contains(c.SkipFields, fieldType.Name) {
			if tag := fieldType.Tag.Get(string(c.Type)); tag != "" {
				keys = append(keys, tag)
			}
		}
	}
	return keys
}

// BuildDevEnv fills the values of the struct with the values from the keychain.
func BuildDevEnv(c *Env, secrets *Secrets, skipFields ...string) error {
	if secrets == nil {
		envKeys := c.GetKeys()
		secrets = GetConfig(c.Name, envKeys...)
	}
	secretMap := secrets.ToMap(skipFields...)

	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()

	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Field(i)
		fieldType := typeOf.Field(i)

		var tag string
		switch c.Type {
		case "json":
			tag = fieldType.Tag.Get("json")
		case "yaml":
			tag = fieldType.Tag.Get("yaml")
		case CUSTOM:
			tag = fieldType.Tag.Get(string(CUSTOM))
		default:
			continue
		}

		if val, ok := secretMap[tag]; ok {
			err := setField(field, val)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %s", fieldType.Name, err)
			}
		}
	}
	return nil
}
