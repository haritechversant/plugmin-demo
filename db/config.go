package db

import (
	"fmt"
)

// DBConfig defines an interface for database configuration.
type DBConfig interface {
	GetDSN() string
}

// PostgresConfig implements DBConfig for PostgreSQL.
type PostgresConfig struct {
	Host     string
	User     string
	DbName   string
	Password string
	Port     int
	SSLMode  string
}

// GetDSN returns the DSN string for PostgreSQL.
func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		p.Host, p.User, p.Password, p.DbName, p.Port, p.SSLMode)
}

// MySQLConfig implements DBConfig for MySQL.
type MySQLConfig struct {
	Host     string
	User     string
	DbName   string
	Password string
	Port     int
}

func (m *MySQLConfig) GetDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DbName)
	return dsn
}
