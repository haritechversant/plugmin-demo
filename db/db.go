package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB initializes a database connection based on the provided config and database type.
func InitDB(config DBConfig, dbType string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	dsn := config.GetDSN()

	switch dbType {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println("Error:", err)
		}
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println("Error:", err)
		}
	default:
		err = fmt.Errorf("unsupported database type: %s", dbType)
	}

	return db, err
}
