package config

import (
	"os"
	"time"
)

type ConfigDB struct {
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func (c ConfigDB) GetDSN() string {
	return os.Getenv("POSTGRES_CONN")
}

func (c ConfigDB) GetMaxOpenConnections() int {
	return c.MaxOpenConns
}

func (c ConfigDB) GetMaxIdleConnections() int {
	return c.MaxIdleConns
}

func (c ConfigDB) GetConnectionMaxLifetime() time.Duration {
	return c.ConnMaxLifetime
}

func (c ConfigDB) GetConnectionMaxIdleTime() time.Duration {
	return c.ConnMaxIdleTime
}
