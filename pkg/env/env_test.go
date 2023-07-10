package env

import (
	"fmt"
	"time"

	"os"
	"testing"

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
	os.Setenv("API_KEY", "abc123")
	os.Setenv("DATABASE_URL", "localhost:5432")

	var appConfig AppConfig
	err := Get(
		PROD,
		&GetConfig{
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
	os.Setenv("API_KEY", "abc123")
	os.Setenv("DATABASE_URL", "localhost:5432")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("DATABASE_URL")
	os.Setenv("CUE_API_KEY", "lol123")
	os.Setenv("CUE_DATABASE_URL", "localhost:9999")
	defer os.Unsetenv("CUE_API_KEY")
	defer os.Unsetenv("CUE_DATABASE_URL")

	var appConfig AppConfig
	err := Get(
		PROD,
		&GetConfig{
			Name:         testJsonfile,
			Type:         "json",
			EnvPrefix:    "CUE",
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
	err := os.WriteFile(testJsonfile, fileContent, 0644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer os.Remove(testJsonfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&GetConfig{
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

	err = os.WriteFile(testYamlfile, yamlData, 0644)
	if err != nil {
		t.Fatalf("failed to create YAML file: %v", err)
	}
	defer os.Remove(testYamlfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&GetConfig{
			Name:         testYamlfile,
			Type:         YAML,
			ConfigStruct: &appConfig,
		},
	)
	fmt.Println(appConfig)
	assert.NoError(t, err)
	assert.Equal(t, "secret-api-key", appConfig.APIKey)
	assert.Equal(t, "https://example.com/db", appConfig.DatabaseURL)
}

func TestInvalidFile(t *testing.T) {
	invalidData := []byte(`{json "invalid": "json"}`)
	err := os.WriteFile(testJsonfile, invalidData, 0644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer os.Remove(testJsonfile)

	var appConfig AppConfig
	err = Get(
		PROD,
		&GetConfig{
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

	config := &GetConfig{
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
	type Config struct {
		AdminGroupID      string `json:"admin_group_id"`
		DBAddress         string `json:"db_address"`
		DBName            string `json:"db_name"`
		DBPass            string `json:"db_pass"`
		DBPort            string `json:"db_port"`
		DBUser            string `json:"db_user"`
		IDPDomain         string `json:"idp_domain"`
		IDPToken          string `json:"idp_token"`
		IDPURL            string `json:"idp_url"`
		MDMKey            string `json:"mdm_key"`
		MDMURL            string `json:"mdm_url"`
		SlackAppToken     string `json:"slack_app_token"`
		SlackAlertChannel string `json:"slack_alert_channel"`
		SlackBotToken     string `json:"slack_bot_token"`
		SlackBotID        string `json:"slack_bot_id"`
	}

	// set the keys
	var cfg Config
	config := &GetConfig{
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

	var cf Config
	err := Get(
		EnvType(DEV),
		&GetConfig{
			Name:         "testService",
			EnvPrefix:    "CUEBERT",
			ConfigStruct: &cf,
			Type:         JSON,
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, "testpassword", cf.AdminGroupID)
	assert.Equal(t, "testpassword", cf.DBAddress)
	assert.Equal(t, "testpassword", cf.DBName)
	assert.Equal(t, "testpassword", cf.DBPass)
	assert.Equal(t, "testpassword", cf.DBPort)
	assert.Equal(t, "testpassword", cf.DBUser)
	assert.Equal(t, "testpassword", cf.IDPDomain)
	assert.Equal(t, "testpassword", cf.IDPToken)
	assert.Equal(t, "testpassword", cf.IDPURL)
	assert.Equal(t, "testpassword", cf.MDMKey)
	assert.Equal(t, "testpassword", cf.MDMURL)
	assert.Equal(t, "testpassword", cf.SlackAppToken)
	assert.Equal(t, "testpassword", cf.SlackAlertChannel)
	assert.Equal(t, "testpassword", cf.SlackBotToken)
	assert.Equal(t, "testpassword", cf.SlackBotID)

	// remove the secrets
	for _, a := range envKeys {
		err := keyring.Delete(a, "testService")
		if err != nil {
			t.Fatal(err)
		}
	}
}
