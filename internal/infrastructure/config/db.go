package config

import (
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/kristianrpo/document-management-microservice/internal/domain"
)

func Connect(dsn string) *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil { log.Fatalf("db connect: %v", err) }
    if err := db.AutoMigrate(&domain.Document{}); err != nil { log.Fatalf("db migrate: %v", err) }
    return db
}