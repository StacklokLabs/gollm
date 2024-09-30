// Copyright 2024 Stacklok, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config interface {
	Get(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// ViperConfig implements the Config interface using Viper.
type ViperConfig struct {
	viper *viper.Viper
}

// NewViperConfig initializes a ViperConfig with a given Viper instance.
func NewViperConfig(v *viper.Viper) *ViperConfig {
	return &ViperConfig{viper: v}
}

// Get returns a string value for the given key.
func (vc *ViperConfig) Get(key string) string {
	return vc.viper.GetString(key)
}

// GetInt returns an integer value for the given key.
func (vc *ViperConfig) GetInt(key string) int {
	return vc.viper.GetInt(key)
}

// GetBool returns a boolean value for the given key.
func (vc *ViperConfig) GetBool(key string) bool {
	return vc.viper.GetBool(key)
}

// InitializeViperConfig initializes and returns a Config implementation using Viper.
// It reads the configuration from the specified config file and paths.
func InitializeViperConfig(configName, configType, configPath string) Config {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(configPath)

	// Read in the config file
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Wrap Viper with ViperConfig and return as Config
	return NewViperConfig(v)
}
