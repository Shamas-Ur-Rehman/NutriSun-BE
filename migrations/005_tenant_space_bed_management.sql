BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'room_type_enum') THEN
        CREATE TYPE room_type_enum AS ENUM ('ICU', 'Ward', 'OT', 'Isolation', 'ER', 'Recovery', 'NICU', 'PICU', 'Lab', 'Office', 'Storage', 'Other');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bed_status_enum') THEN
        CREATE TYPE bed_status_enum AS ENUM ('Available', 'Occupied', 'Maintenance', 'Cleanup', 'Blocked');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_number VARCHAR(50) NOT NULL,
    floor VARCHAR(50) NOT NULL,
    building VARCHAR(100) NOT NULL DEFAULT '',
    facility_id UUID NOT NULL,
    room_type room_type_enum NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Active'
        CHECK (status IN ('Active', 'Inactive')),
    capacity INTEGER NULL,
    is_isolation BOOLEAN NOT NULL DEFAULT FALSE,
    is_negative_pressure BOOLEAN NOT NULL DEFAULT FALSE,
    has_emergency_power BOOLEAN NOT NULL DEFAULT TRUE,
    assigned_nurse_station VARCHAR(100) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT positive_room_capacity CHECK (capacity IS NULL OR capacity > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_location_per_branch
    ON rooms (business_id, branch_id, facility_id, building, floor, room_number);
CREATE INDEX IF NOT EXISTS idx_rooms_facility ON rooms(facility_id);
CREATE INDEX IF NOT EXISTS idx_rooms_type ON rooms(room_type);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);
CREATE INDEX IF NOT EXISTS idx_rooms_isolation ON rooms(is_isolation);
CREATE INDEX IF NOT EXISTS idx_rooms_business_branch_facility ON rooms(business_id, branch_id, facility_id);

CREATE TABLE IF NOT EXISTS beds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bed_number VARCHAR(50) NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    status bed_status_enum NOT NULL DEFAULT 'Available',
    patient_id UUID NULL,
    patient_name VARCHAR(255) NOT NULL DEFAULT '',
    admission_date TIMESTAMPTZ NULL,
    expected_discharge_date DATE NULL,
    discharge_date TIMESTAMPTZ NULL,
    last_cleaned_at TIMESTAMPTZ NULL,
    cleaning_due_at TIMESTAMPTZ NULL,
    isolation_status BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT valid_bed_discharge CHECK (discharge_date IS NULL OR admission_date IS NULL OR discharge_date >= admission_date),
    CONSTRAINT valid_expected_discharge CHECK (expected_discharge_date IS NULL OR admission_date IS NULL OR expected_discharge_date >= DATE(admission_date))
);

ALTER TABLE beds ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
DROP INDEX IF EXISTS uq_beds_room_number_per_branch;
CREATE UNIQUE INDEX IF NOT EXISTS uq_beds_room_number_per_branch
    ON beds (business_id, branch_id, room_id, bed_number)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_beds_room ON beds(room_id);
CREATE INDEX IF NOT EXISTS idx_beds_status ON beds(status);
CREATE INDEX IF NOT EXISTS idx_beds_patient ON beds(patient_id);
CREATE INDEX IF NOT EXISTS idx_beds_cleaning_due ON beds(cleaning_due_at);
CREATE INDEX IF NOT EXISTS idx_beds_business_branch_status ON beds(business_id, branch_id, status);
CREATE INDEX IF NOT EXISTS idx_beds_deleted_at ON beds(deleted_at);

COMMIT;
