package config

import (
	"backend/db"
	"database/sql"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	db *sql.DB
}

func (c Config) GetDB() *sql.DB {
	return c.db
}

func (c Config) GetServerAddress() string {
	return os.Getenv("SERVER_ADDRESS")
}

func NewConfig() *Config {
	type cfg struct {
		PostgresConfig ConfigDB `yaml:"postgres"`
	}
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	var c cfg
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		panic(err)
	}
	conn, err := db.ConnectDB(c.PostgresConfig)
	if err != nil {
		panic(err)
	}
	if err := conn.Ping(); err != nil {
		panic(err)
	}
	return &Config{db: conn}
}
