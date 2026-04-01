package database

import (
	"fmt"
	"log"
	"products/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Client() *gorm.DB {
	return db
}

func Connect() {
	dsn := config.Config().DSN
	if dsn == "" {
		log.Fatal("unable to fund dsn")
	}

	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("unable to connect %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("unable to connect %v", err)
	}
	log.Println("successfully connected to the database")
}

func Disconnect() error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
