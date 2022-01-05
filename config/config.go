/*
 * @file: config.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package config

import (
	"os"
	"strconv"
)

type Config struct {
	Hub struct {
		Addr   string
		Port   string
		Secure bool
	}
	Redis struct {
		Addr   string
		Passwd string
		DB     int
	}
}

// var Config = struct {
// 	Hub struct {
// 		Addr   string
// 		Port   string
// 		Secure bool
// 	}
// 	Redis struct {
// 		Addr   string
// 		Passwd string
// 		DB     int
// 	}
// }{}

func init() {
	hub_addr := os.Getenv("HUB_ADDR")
	if hub_addr == "" {
		Config.Hub.Addr = "localhost"
	} else {
		Config.Hub.Addr = hub_addr
	}
	hub_port := os.Getenv("HUB_PORT")
	if hub_port == "" {
		Config.Hub.Port = "8083"
	} else {
		Config.Hub.Port = hub_port
	}
	if hub_secure, err := strconv.ParseBool(os.Getenv("HUB_SECURE")); err == nil {
		Config.Hub.Secure = hub_secure
	} else {
		Config.Hub.Secure = false
	}
	redis_addr := os.Getenv("REDIS_ADDR")
	if redis_addr == "" {
		Config.Redis.Addr = "localhost:6379"
	} else {
		Config.Redis.Addr = redis_addr
	}
	redis_db := os.Getenv("REDIS_DB")
	if redis_db == "" {
		Config.Redis.DB = 0
	} else {
		db, err := strconv.Atoi(redis_db)
		if err != nil {
			Config.Redis.DB = 0
		} else {
			Config.Redis.DB = db
		}
	}
}

func GetFromEnvOrDefault() *Config {

	config := Config{}
	hub_addr := os.Getenv("HUB_ADDR")
	if hub_addr == "" {
		config.Hub.Addr = "localhost"
	} else {
		config.Hub.Addr = hub_addr
	}
	hub_port := os.Getenv("HUB_PORT")
	if hub_port == "" {
		config.Hub.Port = "8083"
	} else {
		config.Hub.Port = hub_port
	}
	if hub_secure, err := strconv.ParseBool(os.Getenv("HUB_SECURE")); err == nil {
		config.Hub.Secure = hub_secure
	} else {
		config.Hub.Secure = false
	}
	redis_addr := os.Getenv("REDIS_ADDR")
	if redis_addr == "" {
		config.Redis.Addr = "localhost:6379"
	} else {
		config.Redis.Addr = redis_addr
	}
	redis_db := os.Getenv("REDIS_DB")
	if redis_db == "" {
		config.Redis.DB = 0
	} else {
		db, err := strconv.Atoi(redis_db)
		if err != nil {
			config.Redis.DB = 0
		} else {
			config.Redis.DB = db
		}
	}
	return &config
}

func GetDefaultConfig() *Config {

	config := Config{}
	config.Hub.Addr = "localhost"
	config.Hub.Port = "8083"
	config.Hub.Secure = false
	config.Redis.Addr = "localhost:6379"
	config.Redis.DB = 0
	return &config
}

func NewConfig(hub_addr, hub_port string, hub_secure bool, redis_addr string, redis_db int) *Config {
	config := Config{}
	config.Hub.Addr = hub_addr
	config.Hub.Port = hub_port
	config.Hub.Secure = hub_secure
	config.Redis.Addr = redis_addr
	config.Redis.DB = redis_db
	return &config
}
