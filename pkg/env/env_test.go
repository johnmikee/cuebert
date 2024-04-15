package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	APIKey      string `json:"api_key" yaml:"api_key"`
	DatabaseURL string `json:"database_url" yaml:"database_url"`
}

var (
	testJsonfile = ".testconfig.json"
	fileContent  = []byte(`{
		"database_url": "https://example.com/db",
		"api_key": "secret-api-key"
	}`)

	testYamlfile = ".testconfig.yaml"
)

func TestEnvNoPrefix(t *testing.T) {
	t.Setenv("API_KEY", "abc123")
	t.Setenv("DATABASE_URL", "localhost:5432")

	var appConfig AppConfig
	err := Get(
		PROD,
		&Env{
			Name:         testJsonfile,
			Type:         "json",
			ConfigStruct: &appConfig,
		},
	)
	assert.NoError(t, err)

	// Assert values from no prefix env vars
	assert.Equal(t, "abc123", appConfig.APIKey)
	assert.Equal(t, "localhost:5432", appConfig.DatabaseURL)
}

func TestEnvWithPrefix(t *testing.T) {
	t.Setenv("API_KEY", "abc123")
	t.Setenv("DATABASE_URL", "localhost:5432")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("DATABASE_URL")
	t.Setenv("YAE_API_KEY", "lol123")
	t.Setenv("YAE_DATABASE_URL", "localhost:9999")
	defer os.Unsetenv("YAE_API_KEY")
	defer os.Unsetenv("YAE_DATABASE_URL")

	var appConfig AppConfig
	err := Get(
		PROD,
		&Env{
			Name:         testJsonfile,
			Type:         "json",
			EnvPrefix:    "YAE",
			ConfigStruct: &appConfig,
		},
	)
	assert.NoError(t, err)
	// Assert values from prefix env vars
	assert.Equal(t, "lol123", appConfig.APIKey)
	assert.Equal(t, "localhost:9999", appConfig.DatabaseURL)
	// Make sure we don't use the non-prefixed env vars
	assert.NotEqual(t, "abc123", appConfig.APIKey)
	assert.NotEqual(t, "localhost:5432", appConfig.DatabaseURL)
}

func TestJSON(t *testing.T) {
	err := os.WriteFile(testJsonfile, fileContent, 0o644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer os.Remove(testJsonfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&Env{
			Name:         testJsonfile,
			Type:         JSON,
			ConfigStruct: &appConfig,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "secret-api-key", appConfig.APIKey)
	assert.Equal(t, "https://example.com/db", appConfig.DatabaseURL)
}

func TestYAML(t *testing.T) {
	data := map[string]string{
		"database_url": "https://example.com/db",
		"api_key":      "secret-api-key",
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data to YAML: %v", err)
	}

	err = os.WriteFile(testYamlfile, yamlData, 0o644)
	if err != nil {
		t.Fatalf("failed to create YAML file: %v", err)
	}
	defer os.Remove(testYamlfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&Env{
			Name:         testYamlfile,
			Type:         YAML,
			ConfigStruct: &appConfig,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "secret-api-key", appConfig.APIKey)
	assert.Equal(t, "https://example.com/db", appConfig.DatabaseURL)
}

func TestInvalidFile(t *testing.T) {
	invalidData := []byte(`{json "invalid": "json"}`)
	err := os.WriteFile(testJsonfile, invalidData, 0o644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer os.Remove(testJsonfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&Env{
			Name:         testJsonfile,
			Type:         "json",
			ConfigStruct: appConfig,
		},
	)
	assert.Error(t, err)
}

func TestDevEnvSetKeys(t *testing.T) {
	type AppConfig struct {
		APIKey      string `json:"api_key"`
		DatabaseURL string `json:"database_url"`
	}

	var appConfig AppConfig

	type Secrets struct {
		Name  string
		Value string
	}
	secrets := []Secrets{}

	config := &Env{
		Name:         "test",
		Type:         JSON,
		Path:         ".",
		EnvPrefix:    "prefixed",
		ConfigStruct: &appConfig,
	}
	envKeys := config.GetKeys()

	for _, a := range envKeys {
		err := keyring.Set("test", a, "test")
		if err != nil {
			t.Fatal(err)
		}

		secret, err := keyring.Get("test", a)
		if err != nil {
			t.Fatal(err)
		}

		if secret != "test" {
			t.Fatalf("expected %s, got %s", "test", secret)
		}

		secrets = append(secrets, Secrets{Name: a, Value: secret})
	}

	for _, s := range secrets {
		switch s.Name {
		case "api_key":
			appConfig.APIKey = s.Value
		case "database_url":
			appConfig.DatabaseURL = s.Value
		}
	}

	// Assert the values
	assert.Equal(t, "test", appConfig.APIKey)
	assert.Equal(t, "test", appConfig.DatabaseURL)

	// Additional assertions
	assert.NotEmpty(t, appConfig.APIKey)
	assert.NotEmpty(t, appConfig.DatabaseURL)
	assert.Len(t, secrets, 2)
}

func TestDevEnvBiggerStruct(t *testing.T) {
	type Conf struct {
		DBAddress string `json:"db_address"`
		DBName    string `json:"db_name"`
		DBPass    string `json:"db_pass"`
		DBPort    string `json:"db_port"`
		DBUser    string `json:"db_user"`
	}

	// set the keys
	var cfg Conf
	config := &Env{
		Name:         "test",
		Type:         JSON,
		Path:         ".",
		EnvPrefix:    "prefixed",
		ConfigStruct: &cfg,
	}

	envKeys := config.GetKeys()

	for _, a := range envKeys {
		err := keyring.Set(a, "testService", "testpassword")
		if err != nil {
			t.Fatal(err)
		}
	}

	// there is some small delay in getting the keys
	time.Sleep(1 * time.Second)

	var cf Conf
	err := Get(
		EnvType(DEV),
		&Env{
			Name:         "testService",
			EnvPrefix:    "YAE",
			ConfigStruct: &cf,
			Type:         JSON,
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, "testpassword", cf.DBAddress)
	assert.Equal(t, "testpassword", cf.DBName)
	assert.Equal(t, "testpassword", cf.DBPass)
	assert.Equal(t, "testpassword", cf.DBPort)
	assert.Equal(t, "testpassword", cf.DBUser)

	// remove the secrets
	for _, a := range envKeys {
		err := keyring.Delete(a, "testService")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSkipFields(t *testing.T) {
	type Conf struct {
		DBAddress string `json:"db_address"`
		DBName    string `json:"db_name"`
		DBPass    string `json:"db_pass"`
		DBPort    string `json:"db_port"`
		DBUser    string `json:"db_user"`
	}

	// set the keys
	var cfg Conf
	config := &Env{
		Name:         "test",
		Type:         JSON,
		Path:         ".",
		EnvPrefix:    "prefixed",
		ConfigStruct: &cfg,
	}

	envKeys := config.GetKeys()

	for _, a := range envKeys {
		err := keyring.Set(a, "testService", "testpassword")
		if err != nil {
			t.Fatal(err)
		}
	}

	// there is some small delay in getting the keys
	time.Sleep(1 * time.Second)

	var cf Conf
	err := Get(
		EnvType(DEV),
		&Env{
			Name:         "testService",
			EnvPrefix:    "YAE",
			ConfigStruct: &cf,
			Type:         JSON,
			SkipFields:   []string{"DBPass", "DBPort"},
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, "testpassword", cf.DBAddress)
	assert.Equal(t, "testpassword", cf.DBName)
	assert.Equal(t, "", cf.DBPass)
	assert.Equal(t, "", cf.DBPort)
	assert.Equal(t, "testpassword", cf.DBUser)

	// remove the secrets
	for _, a := range envKeys {
		err := keyring.Delete(a, "testService")
		if err != nil {
			t.Fatal(err)
		}
	}
}
