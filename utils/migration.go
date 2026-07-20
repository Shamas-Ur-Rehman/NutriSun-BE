package utils

import (
	"fmt"
	"log"
	"strings"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/models"

	"gorm.io/gorm"
)

func AddPermissionCompositeIndex(db *gorm.DB) error {
	var indexExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = current_schema()
			AND tablename = 'permissions'
			AND indexname = 'idx_role_module_action'
		)
	`).Scan(&indexExists).Error

	if err != nil {
		return err
	}
	if !indexExists {
		return db.Exec(`
			CREATE INDEX idx_role_module_action 
			ON permissions(role_id, module_id, action)
		`).Error
	}

	return nil
}

func RunTenantFacilityOperationsMigrations(masterDB *gorm.DB) error {
	var businesses []models.Business
	if err := masterDB.Select("id", "name_en", "db").Find(&businesses).Error; err != nil {
		return fmt.Errorf("failed to load tenant databases: %w", err)
	}

	seen := make(map[string]struct{})
	for _, business := range businesses {
		dbName := strings.TrimSpace(business.DB)
		if dbName == "" {
			dbName = GetDefaultTenantDBName()
		}
		if _, ok := seen[dbName]; ok {
			continue
		}
		seen[dbName] = struct{}{}

		tenantDB := config.ConnectTenantSQLDB(dbName)
		if tenantDB == nil {
			return fmt.Errorf("failed to connect tenant database %q", dbName)
		}

		if err := RunFacilityOperationsMigration(tenantDB); err != nil {
			return fmt.Errorf("failed to migrate tenant database %q: %w", dbName, err)
		}
		log.Printf("tenant facility operations migration ready for %s", dbName)
	}

	return nil
}

func RunTenantBedSoftDeleteMigrations(masterDB *gorm.DB) error {
	var businesses []models.Business
	if err := masterDB.Select("id", "name_en", "db").Find(&businesses).Error; err != nil {
		return fmt.Errorf("failed to load tenant databases: %w", err)
	}

	seen := make(map[string]struct{})
	for _, business := range businesses {
		dbName := strings.TrimSpace(business.DB)
		if dbName == "" {
			dbName = GetDefaultTenantDBName()
		}
		if _, ok := seen[dbName]; ok {
			continue
		}
		seen[dbName] = struct{}{}

		tenantDB := config.ConnectTenantSQLDB(dbName)
		if tenantDB == nil {
			return fmt.Errorf("failed to connect tenant database %q", dbName)
		}

		if err := RunBedSoftDeleteMigration(tenantDB); err != nil {
			return fmt.Errorf("failed to migrate tenant database %q: %w", dbName, err)
		}
		log.Printf("tenant bed soft-delete migration ready for %s", dbName)
	}

	return nil
}

func RunBedSoftDeleteMigration(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE beds ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL`,
		`DROP INDEX IF EXISTS uq_beds_room_number_per_branch`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_beds_room_number_per_branch
			ON beds (business_id, branch_id, room_id, bed_number)
			WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_beds_deleted_at ON beds(deleted_at)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}

func RunFacilityOperationsMigration(db *gorm.DB) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'waste_type_enum') THEN
				CREATE TYPE waste_type_enum AS ENUM ('Infectious', 'Sharps', 'Chemical', 'General');
			END IF;
		END $$`,
		`CREATE TABLE IF NOT EXISTS housekeeping_tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			task_number VARCHAR(50) NOT NULL,
			room_id UUID NOT NULL,
			task_type VARCHAR(100) NOT NULL,
			frequency VARCHAR(50) NOT NULL,
			assigned_to_staff BIGINT NULL,
			assigned_to_team VARCHAR(100) NOT NULL DEFAULT '',
			scheduled_date DATE NOT NULL,
			scheduled_time TIME NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'Scheduled'
				CHECK (status IN ('Scheduled', 'In Progress', 'Completed')),
			completed_at TIMESTAMPTZ NULL,
			supervisor_notes TEXT NOT NULL DEFAULT '',
			cleaning_checklist JSONB NOT NULL DEFAULT '[]'::JSONB,
			signature_url TEXT NOT NULL DEFAULT '',
			is_covid_zone BOOLEAN NOT NULL DEFAULT FALSE,
			requires_ppe BOOLEAN NOT NULL DEFAULT FALSE,
			business_id BIGINT NOT NULL,
			branch_id BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by BIGINT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by BIGINT NOT NULL,
			CONSTRAINT valid_housekeeping_completion CHECK (completed_at IS NULL OR completed_at >= created_at)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_housekeeping_task_number_per_branch
			ON housekeeping_tasks (business_id, branch_id, task_number)`,
		`CREATE INDEX IF NOT EXISTS idx_housekeeping_room ON housekeeping_tasks(room_id)`,
		`CREATE INDEX IF NOT EXISTS idx_housekeeping_status ON housekeeping_tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_housekeeping_scheduled_date ON housekeeping_tasks(scheduled_date)`,
		`CREATE INDEX IF NOT EXISTS idx_housekeeping_business_branch_date
			ON housekeeping_tasks(business_id, branch_id, scheduled_date)`,
		`CREATE TABLE IF NOT EXISTS waste_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			waste_number VARCHAR(50) NOT NULL,
			waste_type waste_type_enum NOT NULL,
			quantity_kg DECIMAL(10,2) NOT NULL,
			container_type VARCHAR(50) NOT NULL DEFAULT '',
			container_count INTEGER NULL,
			origin_department VARCHAR(100) NOT NULL DEFAULT '',
			origin_room_id UUID NULL,
			disposed_by_vendor_id UUID NULL,
			disposal_date DATE NOT NULL,
			disposal_time TIME NULL,
			manifest_number VARCHAR(100) NOT NULL,
			transporter_name VARCHAR(255) NOT NULL DEFAULT '',
			destination_facility_id UUID NULL,
			destination_facility VARCHAR(255) NOT NULL DEFAULT '',
			certificate_url TEXT NOT NULL DEFAULT '',
			authorized_by BIGINT NOT NULL,
			notes TEXT NOT NULL DEFAULT '',
			business_id BIGINT NOT NULL,
			branch_id BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by BIGINT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by BIGINT NOT NULL,
			CONSTRAINT positive_quantity CHECK (quantity_kg > 0),
			CONSTRAINT positive_container_count CHECK (container_count IS NULL OR container_count >= 0)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_waste_number_per_branch
			ON waste_records (business_id, branch_id, waste_number)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_waste_manifest_per_branch
			ON waste_records (business_id, branch_id, manifest_number)`,
		`CREATE INDEX IF NOT EXISTS idx_waste_records_disposal_date ON waste_records(disposal_date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_waste_records_waste_type ON waste_records(waste_type)`,
		`CREATE INDEX IF NOT EXISTS idx_waste_records_manifest ON waste_records(manifest_number)`,
		`ALTER TABLE waste_records ADD COLUMN IF NOT EXISTS destination_facility_id UUID NULL`,
		`CREATE INDEX IF NOT EXISTS idx_waste_records_destination_facility ON waste_records(destination_facility_id)`,
		`CREATE INDEX IF NOT EXISTS idx_waste_records_business_branch_date
			ON waste_records(business_id, branch_id, disposal_date DESC)`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
