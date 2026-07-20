BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'waste_type_enum') THEN
        CREATE TYPE waste_type_enum AS ENUM ('Infectious', 'Sharps', 'Chemical', 'General');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS housekeeping_tasks (
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
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_housekeeping_task_number_per_branch
    ON housekeeping_tasks (business_id, branch_id, task_number);
CREATE INDEX IF NOT EXISTS idx_housekeeping_room ON housekeeping_tasks(room_id);
CREATE INDEX IF NOT EXISTS idx_housekeeping_status ON housekeeping_tasks(status);
CREATE INDEX IF NOT EXISTS idx_housekeeping_scheduled_date ON housekeeping_tasks(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_housekeeping_business_branch_date
    ON housekeeping_tasks(business_id, branch_id, scheduled_date);

CREATE TABLE IF NOT EXISTS waste_records (
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
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_waste_number_per_branch
    ON waste_records (business_id, branch_id, waste_number);
CREATE UNIQUE INDEX IF NOT EXISTS uq_waste_manifest_per_branch
    ON waste_records (business_id, branch_id, manifest_number);
CREATE INDEX IF NOT EXISTS idx_waste_records_disposal_date ON waste_records(disposal_date DESC);
CREATE INDEX IF NOT EXISTS idx_waste_records_waste_type ON waste_records(waste_type);
CREATE INDEX IF NOT EXISTS idx_waste_records_manifest ON waste_records(manifest_number);
CREATE INDEX IF NOT EXISTS idx_waste_records_destination_facility ON waste_records(destination_facility_id);
CREATE INDEX IF NOT EXISTS idx_waste_records_business_branch_date
    ON waste_records(business_id, branch_id, disposal_date DESC);

COMMIT;
