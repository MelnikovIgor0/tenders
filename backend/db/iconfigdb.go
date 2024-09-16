package db

import "time"

type IConfigDB interface {
	GetDSN() string
	GetMaxOpenConnections() int
	GetMaxIdleConnections() int
	GetConnectionMaxLifetime() time.Duration
	GetConnectionMaxIdleTime() time.Duration
}
