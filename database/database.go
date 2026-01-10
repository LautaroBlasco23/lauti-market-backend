package database

import "database/sql"

type Config struct {
	Driver string
	DSN    string
}

type Database interface {
	DB() *sql.DB
	Close() error
	Ping() error
}
