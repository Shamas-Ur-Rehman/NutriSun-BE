package utils

import (
	"fmt"
	"log"
	"strings"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/models"

	"gorm.io/gorm"
)

func AutoMigrateMasterDB(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Business{},
		&models.Subscription{},
		&models.Branch{},
		&models.Role{},
		&models.Module{},
		&models.Permission{},
		&models.User{},
	); err != nil {
		return fmt.Errorf("auto-migrate master schema: %w", err)
	}

	if err := AddPermissionCompositeIndex(db); err != nil {
		return fmt.Errorf("ensure permission composite index: %w", err)
	}

	return nil
}

func AutoMigrateTenantDatabases(masterDB *gorm.DB) error {
	dbNames, err := loadTenantDatabaseNames(masterDB)
	if err != nil {
		return err
	}

	for _, dbName := range dbNames {
		tenantDB := config.ConnectTenantSQLDB(dbName)
		if tenantDB == nil {
			return fmt.Errorf("failed to connect tenant database %q", dbName)
		}

		if err := AutoMigrateTenantDB(tenantDB); err != nil {
			return fmt.Errorf("failed to auto-migrate tenant database %q: %w", dbName, err)
		}

		log.Printf("tenant schema ready for %s", dbName)
	}

	return nil
}

func AutoMigrateTenantDB(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Customer{},
		&models.CustomerAddress{},
		&models.MenuMonth{},
		&models.MenuDay{},
		&models.MenuDayItem{},
		&models.SubscriptionPlan{},
		&models.CustomerSubscription{},
		&models.CustomerSubscriptionDay{},
	); err != nil {
		return fmt.Errorf("auto-migrate tenant schema: %w", err)
	}

	return nil
}

func loadTenantDatabaseNames(masterDB *gorm.DB) ([]string, error) {
	var businesses []models.Business
	if err := masterDB.Select("id", "name_en", "db").Find(&businesses).Error; err != nil {
		return nil, fmt.Errorf("failed to load tenant databases: %w", err)
	}

	seen := make(map[string]struct{})
	dbNames := make([]string, 0, len(businesses)+1)

	appendDBName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, exists := seen[name]; exists {
			return
		}

		seen[name] = struct{}{}
		dbNames = append(dbNames, name)
	}

	for _, business := range businesses {
		dbName := business.DB
		if strings.TrimSpace(dbName) == "" {
			dbName = GetDefaultTenantDBName()
		}
		appendDBName(dbName)
	}

	if len(dbNames) == 0 {
		appendDBName(GetDefaultTenantDBName())
	}

	return dbNames, nil
}
