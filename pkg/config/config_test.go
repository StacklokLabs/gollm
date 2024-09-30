// config_test.go
package config

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestViperConfig_Get(t *testing.T) {
	// Create a new Viper instance and set a string value
	v := viper.New()
	v.Set("stringKey", "stringValue")

	// Initialize ViperConfig with the Viper instance
	vc := NewViperConfig(v)

	// Test the Get method
	value := vc.Get("stringKey")
	if value != "stringValue" {
		t.Errorf("Expected 'stringValue', got '%s'", value)
	}
}

func TestViperConfig_GetInt(t *testing.T) {
	// Create a new Viper instance and set an integer value
	v := viper.New()
	v.Set("intKey", 42)

	// Initialize ViperConfig with the Viper instance
	vc := NewViperConfig(v)

	// Test the GetInt method
	value := vc.GetInt("intKey")
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

func TestViperConfig_GetBool(t *testing.T) {
	// Create a new Viper instance and set a boolean value
	v := viper.New()
	v.Set("boolKey", true)

	// Initialize ViperConfig with the Viper instance
	vc := NewViperConfig(v)

	// Test the GetBool method
	value := vc.GetBool("boolKey")
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}
}

func TestInitializeViperConfig(t *testing.T) {
	// Since InitializeViperConfig reads from a file, we'll create a temporary config file for testing
	configName := "testconfig"
	configType := "yaml"
	configPath := "."

	// Create a temporary config file with some test data
	testConfigContent := `
stringKey: stringValue
intKey: 42
boolKey: true
`
	// Write the test config content to a temporary file
	configFileName := configName + "." + configType
	err := writeTempConfigFile(configFileName, testConfigContent)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}
	defer removeTempConfigFile(configFileName)

	// Initialize the config
	cfg := InitializeViperConfig(configName, configType, configPath)

	// Test the values
	if cfg.Get("stringKey") != "stringValue" {
		t.Errorf("Expected 'stringValue', got '%s'", cfg.Get("stringKey"))
	}
	if cfg.GetInt("intKey") != 42 {
		t.Errorf("Expected 42, got %d", cfg.GetInt("intKey"))
	}
	if cfg.GetBool("boolKey") != true {
		t.Errorf("Expected true, got %v", cfg.GetBool("boolKey"))
	}
}

func writeTempConfigFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func removeTempConfigFile(filename string) {
	os.Remove(filename)
}
