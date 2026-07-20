package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	tenantDBCache   sync.Map
	tenantDBCacheMu sync.Mutex
)

func ConnectMasterDB() *gorm.DB {
	db, err := openPostgresDB(resolvePostgresConfig("DB_", os.Getenv("DB_NAME")))
	if err != nil {
		log.Printf("cannot connect master DB: %v", err)
		return nil
	}

	return configureSQLDB(db, poolConfig{
		MaxIdleConns:    getEnvInt(5, "DB_MAX_IDLE_CONNS", "MASTER_DB_MAX_IDLE_CONNS"),
		MaxOpenConns:    getEnvInt(10, "DB_MAX_OPEN_CONNS", "MASTER_DB_MAX_OPEN_CONNS"),
		ConnMaxLifetime: getEnvDuration(30*time.Minute, "DB_CONN_MAX_LIFETIME", "MASTER_DB_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: getEnvDuration(10*time.Minute, "DB_CONN_MAX_IDLE_TIME", "MASTER_DB_CONN_MAX_IDLE_TIME"),
	})
}

func ConnectTenantSQLDB(dbName string) *gorm.DB {
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
	}

	if cachedDB, ok := tenantDBCache.Load(dbName); ok {
		return cachedDB.(*gorm.DB)
	}

	tenantDBCacheMu.Lock()
	defer tenantDBCacheMu.Unlock()

	if cachedDB, ok := tenantDBCache.Load(dbName); ok {
		return cachedDB.(*gorm.DB)
	}

	db, err := openPostgresDB(resolvePostgresConfig("TENANT_DB_", dbName))
	if err != nil {
		log.Printf("cannot connect tenant DB %q: %v", dbName, err)
		return nil
	}

	configuredDB := configureSQLDB(db, poolConfig{
		MaxIdleConns:    getEnvInt(2, "TENANT_DB_MAX_IDLE_CONNS", "DB_MAX_IDLE_CONNS"),
		MaxOpenConns:    getEnvInt(5, "TENANT_DB_MAX_OPEN_CONNS", "DB_MAX_OPEN_CONNS"),
		ConnMaxLifetime: getEnvDuration(30*time.Minute, "TENANT_DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: getEnvDuration(10*time.Minute, "TENANT_DB_CONN_MAX_IDLE_TIME", "DB_CONN_MAX_IDLE_TIME"),
	})
	if configuredDB == nil {
		return nil
	}

	tenantDBCache.Store(dbName, configuredDB)
	return configuredDB
}

func resolvePostgresConfig(prefix, defaultDBName string) postgresConfig {
	return postgresConfig{
		Host:     fallbackEnv(prefix+"HOST", "DB_HOST"),
		Port:     fallbackEnv(prefix+"PORT", "DB_PORT"),
		User:     fallbackEnv(prefix+"USER", "DB_USER"),
		Password: firstNonEmpty(os.Getenv(prefix+"PASSWORD"), os.Getenv(prefix+"PASS"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PASS")),
		Name:     firstNonEmpty(os.Getenv(prefix+"NAME"), defaultDBName),
		SSLMode:  firstNonEmpty(os.Getenv(prefix+"SSLMODE"), os.Getenv("DB_SSLMODE"), "disable"),
		Timezone: firstNonEmpty(os.Getenv(prefix+"TIMEZONE"), os.Getenv("DB_TIMEZONE"), "UTC"),
	}
}

type postgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

func openPostgresDB(cfg postgresConfig) (*gorm.DB, error) {
	if cfg.Port == "" {
		cfg.Port = "5432"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.Timezone,
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		QueryFields:            true,
	})
}

type poolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func configureSQLDB(db *gorm.DB, cfg poolConfig) *gorm.DB {
	connPool, err := db.DB()
	if err != nil {
		log.Printf("failed to get database instance: %v", err)
		return nil
	}

	connPool.SetMaxIdleConns(cfg.MaxIdleConns)
	connPool.SetMaxOpenConns(cfg.MaxOpenConns)
	connPool.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	connPool.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db
}

func fallbackEnv(primary, secondary string) string {
	return firstNonEmpty(os.Getenv(primary), os.Getenv(secondary))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func getEnvInt(defaultValue int, keys ...string) int {
	for _, key := range keys {
		rawValue := os.Getenv(key)
		if rawValue == "" {
			continue
		}

		value, err := strconv.Atoi(rawValue)
		if err != nil {
			log.Printf("invalid integer for %s: %q", key, rawValue)
			continue
		}

		return value
	}

	return defaultValue
}

func getEnvDuration(defaultValue time.Duration, keys ...string) time.Duration {
	for _, key := range keys {
		rawValue := os.Getenv(key)
		if rawValue == "" {
			continue
		}

		value, err := time.ParseDuration(rawValue)
		if err != nil {
			log.Printf("invalid duration for %s: %q", key, rawValue)
			continue
		}

		return value
	}

	return defaultValue
}
