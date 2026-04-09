package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // postgres driver
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

type Postgres struct {
	db *sql.DB
}

func NewPostgres(cfg *PostgresConfig) (*Postgres, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) Ping() error {
	return p.db.Ping()
}
