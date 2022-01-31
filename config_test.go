/*
 * @file: config_test.go
 * @author: Jorge Quitério
 * @copyright (c) 2022 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"os"
	"testing"
)

func TestGetFromEnvOrDefault(t *testing.T) {

	os.Setenv("HUB_ADDR", "localhost")
	os.Setenv("HUB_PORT", "8083")
	os.Setenv("REDIS_DB", "0")

	defaultConfig := GetFromEnvOrDefault()
	if defaultConfig.Hub.Addr != "localhost" {
		t.Errorf("Expected 'localhost' but got %s", defaultConfig.Hub.Addr)
	}
	if defaultConfig.Hub.Port != "8083" {
		t.Errorf("Expected '8083' but got %s", defaultConfig.Hub.Port)
	}
	if defaultConfig.Hub.Secure != false {
		t.Errorf("Expected 'false' but got %t", defaultConfig.Hub.Secure)
	}
	if defaultConfig.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected 'localhost:6379' but got %s", defaultConfig.Redis.Addr)
	}
	if defaultConfig.Redis.DB != 0 {
		t.Errorf("Expected '0' but got %d", defaultConfig.Redis.DB)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	if config.Hub.Addr != "localhost" {
		t.Errorf("Expected 'localhost' but got %s", config.Hub.Addr)
	}
	if config.Hub.Port != "8083" {
		t.Errorf("Expected '8083' but got %s", config.Hub.Port)
	}
	if config.Hub.Secure != false {
		t.Errorf("Expected 'false' but got %t", config.Hub.Secure)
	}
	if config.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected 'localhost:6379' but got %s", config.Redis.Addr)
	}
	if config.Redis.DB != 0 {
		t.Errorf("Expected '0' but got %d", config.Redis.DB)
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig("localhost", "8083", false, "localhost:6379", 0)
	if config.Hub.Addr != "localhost" {
		t.Errorf("Expected 'localhost' but got %s", config.Hub.Addr)
	}
	if config.Hub.Port != "8083" {
		t.Errorf("Expected '8083' but got %s", config.Hub.Port)
	}
	if config.Hub.Secure != false {
		t.Errorf("Expected 'false' but got %t", config.Hub.Secure)
	}
	if config.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected 'localhost:6379' but got %s", config.Redis.Addr)
	}
	if config.Redis.DB != 0 {
		t.Errorf("Expected '0' but got %d", config.Redis.DB)
	}
}
