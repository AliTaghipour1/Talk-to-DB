package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type TalkToDBConfig struct {
	DebugMode      bool
	CliBot         Bot
	Databases      []Database
	AllowedUserIds []int64
}
type Driver string

const (
	Postgres Driver = "postgres"
	MySQL    Driver = "mysql"
)

type Database struct {
	Host   string
	Port   string
	User   string
	Pass   string
	Name   string
	Driver Driver
}

type Bot struct {
	Token string
}

func NewTalkToDbConfig() *TalkToDBConfig {
	config := os.Getenv("config")
	var result TalkToDBConfig
	err := json.Unmarshal([]byte(config), &result)
	if err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}

	return &result
}
