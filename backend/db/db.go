package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func ConnectDB(cfg IConfigDB) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.GetDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.GetMaxOpenConnections())
	db.SetMaxIdleConns(cfg.GetMaxIdleConnections())
	db.SetConnMaxLifetime(cfg.GetConnectionMaxLifetime())
	db.SetConnMaxIdleTime(cfg.GetConnectionMaxIdleTime())
	return db, nil
}
