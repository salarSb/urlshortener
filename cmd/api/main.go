package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/salarSb/urlshortener/internal/shortener"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	BaseURL    string
	ServerPort string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "urlshortener"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "urlshortener"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func main() {
	cfg := loadConfig()
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {
			log.Fatalf("failed to close db: %v", err)
		}
	}(sqlDB)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	repo := shortener.NewRepository(db)
	if err := repo.Migrate(context.Background()); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	handler := shortener.NewHandler(repo, cfg.BaseURL)
	handler.RegisterRoutes(r)
	addr := ":" + cfg.ServerPort
	log.Printf("URL shortener listening on %s (base URL: %s)", addr, cfg.BaseURL)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
