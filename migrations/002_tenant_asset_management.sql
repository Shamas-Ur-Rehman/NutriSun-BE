BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    parent_category_id UUID NULL REFERENCES asset_categories(id) ON DELETE SET NULL,
    calibration_required BOOLEAN NOT NULL DEFAULT FALSE,
    calibration_frequency_days INTEGER NULL,
    warranty_period_days INTEGER NULL,
    expected_lifespan_years INTEGER NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_asset_categories_code_per_branch
    ON asset_categories (business_id, branch_id, LOWER(code));
CREATE UNIQUE INDEX IF NOT EXISTS uq_asset_categories_name_per_branch
    ON asset_categories (business_id, branch_id, LOWER(name));
CREATE INDEX IF NOT EXISTS idx_asset_categories_parent ON asset_categories(parent_category_id);
CREATE INDEX IF NOT EXISTS idx_asset_categories_business_branch ON asset_categories(business_id, branch_id);

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_tag VARCHAR(50) NOT NULL,
    barcode VARCHAR(100) NULL,
    qr_code_url TEXT NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category_id UUID NULL REFERENCES asset_categories(id) ON DELETE SET NULL,
    category VARCHAR(255) NOT NULL,
    sub_category VARCHAR(100) NOT NULL DEFAULT '',
    manufacturer VARCHAR(255) NOT NULL DEFAULT '',
    model VARCHAR(255) NOT NULL DEFAULT '',
    serial_number VARCHAR(255) NULL,
    facility_id UUID NOT NULL,
    facility_name VARCHAR(255) NOT NULL DEFAULT '',
    floor VARCHAR(50) NOT NULL DEFAULT '',
    room_id UUID NULL,
    room_name VARCHAR(100) NOT NULL DEFAULT '',
    zone VARCHAR(100) NOT NULL DEFAULT '',
    gps_coordinates VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'Operational'
        CHECK (status IN ('Operational', 'Under Maintenance', 'Under Repair', 'Retired', 'Disposed', 'Inactive')),
    lifecycle_stage VARCHAR(50) NOT NULL DEFAULT 'In Service'
        CHECK (lifecycle_stage IN ('Procurement', 'In Service', 'Under Repair', 'Retired', 'Disposed')),
    purchase_date DATE NULL,
    purchase_cost DECIMAL(12,2) NULL,
    supplier_id UUID NULL,
    supplier_name VARCHAR(255) NOT NULL DEFAULT '',
    po_number VARCHAR(100) NOT NULL DEFAULT '',
    invoice_number VARCHAR(100) NOT NULL DEFAULT '',
    warranty_start_date DATE NULL,
    warranty_end_date DATE NULL,
    warranty_type VARCHAR(50) NOT NULL DEFAULT '',
    warranty_provider VARCHAR(255) NOT NULL DEFAULT '',
    warranty_contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    calibration_required BOOLEAN NOT NULL DEFAULT FALSE,
    calibration_frequency_days INTEGER NULL,
    last_calibration_date DATE NULL,
    last_calibration_by BIGINT NULL,
    next_calibration_due DATE NULL,
    calibration_certificate_url TEXT NOT NULL DEFAULT '',
    last_service_date DATE NULL,
    last_service_type VARCHAR(50) NOT NULL DEFAULT '',
    total_downtime_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
    mtbf_days DECIMAL(10,2) NULL,
    mttr_hours DECIMAL(10,2) NULL,
    assigned_to_user_id BIGINT NULL,
    assigned_to_department VARCHAR(100) NOT NULL DEFAULT '',
    assigned_date DATE NULL,
    documents JSONB NOT NULL DEFAULT '[]'::JSONB,
    custom_attributes JSONB NOT NULL DEFAULT '{}'::JSONB,
    primary_image_url TEXT NOT NULL DEFAULT '',
    additional_images JSONB NOT NULL DEFAULT '[]'::JSONB,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT valid_warranty_dates CHECK (warranty_end_date IS NULL OR warranty_start_date IS NULL OR warranty_end_date >= warranty_start_date),
    CONSTRAINT valid_calibration_due CHECK (next_calibration_due IS NULL OR last_calibration_date IS NULL OR next_calibration_due >= last_calibration_date),
    CONSTRAINT positive_purchase_cost CHECK (purchase_cost IS NULL OR purchase_cost >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_assets_asset_tag_per_branch
    ON assets (business_id, branch_id, asset_tag) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_assets_serial_number_per_branch
    ON assets (business_id, branch_id, serial_number) WHERE deleted_at IS NULL AND serial_number IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_assets_barcode_per_branch
    ON assets (business_id, branch_id, barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
CREATE INDEX IF NOT EXISTS idx_assets_facility_id ON assets(facility_id);
CREATE INDEX IF NOT EXISTS idx_assets_room_id ON assets(room_id);
CREATE INDEX IF NOT EXISTS idx_assets_assigned_to ON assets(assigned_to_user_id);
CREATE INDEX IF NOT EXISTS idx_assets_calibration_due ON assets(next_calibration_due) WHERE calibration_required = TRUE;
CREATE INDEX IF NOT EXISTS idx_assets_warranty_end ON assets(warranty_end_date);
CREATE INDEX IF NOT EXISTS idx_assets_created_at ON assets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assets_business_branch_status ON assets(business_id, branch_id, status);
CREATE INDEX IF NOT EXISTS idx_assets_business_branch_category ON assets(business_id, branch_id, category);
CREATE INDEX IF NOT EXISTS idx_assets_search ON assets USING GIN(
    to_tsvector('english', name || ' ' || COALESCE(description, '') || ' ' || COALESCE(manufacturer, '') || ' ' || COALESCE(serial_number, ''))
);

CREATE TABLE IF NOT EXISTS asset_calibration_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    calibration_date DATE NOT NULL,
    certificate_url TEXT NOT NULL DEFAULT '',
    next_calibration_due DATE NULL,
    performed_by BIGINT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_asset_calibration_records_asset_id ON asset_calibration_records(asset_id, calibration_date DESC);
CREATE INDEX IF NOT EXISTS idx_asset_calibration_records_business_branch ON asset_calibration_records(business_id, branch_id);

CREATE TABLE IF NOT EXISTS asset_maintenance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    work_order_reference VARCHAR(100) NOT NULL DEFAULT '',
    maintenance_type VARCHAR(50) NOT NULL DEFAULT '',
    maintenance_date DATE NOT NULL,
    technician_name VARCHAR(255) NOT NULL DEFAULT '',
    technician_user_id BIGINT NULL,
    description TEXT NOT NULL DEFAULT '',
    downtime_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
    cost DECIMAL(12,2) NOT NULL DEFAULT 0,
    parts_replaced JSONB NOT NULL DEFAULT '[]'::JSONB,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_asset_maintenance_records_asset_id ON asset_maintenance_records(asset_id, maintenance_date DESC);
CREATE INDEX IF NOT EXISTS idx_asset_maintenance_records_business_branch ON asset_maintenance_records(business_id, branch_id);

INSERT INTO asset_categories (
    code, name, description, calibration_required, calibration_frequency_days,
    warranty_period_days, expected_lifespan_years, is_active,
    business_id, branch_id, created_by, updated_by
)
VALUES
    ('MED-EQ', 'Medical Equipment', 'Medical diagnostic and treatment equipment', TRUE, 365, 1095, 10, TRUE, 1, 1, 1, 1),
    ('FURN', 'Furniture', 'Healthcare furniture and fixtures', FALSE, NULL, 365, 7, TRUE, 1, 1, 1, 1),
    ('IT', 'IT Infrastructure', 'Servers, network devices, and workstations', FALSE, NULL, 365, 5, TRUE, 1, 1, 1, 1)
ON CONFLICT DO NOTHING;

COMMIT;
