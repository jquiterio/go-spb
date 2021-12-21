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

var Config = struct {
	Hub struct {
		Addr string
		Port string
	}
	Redis struct {
		Addr string
		DB   int
	}
}{}

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
