BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'work_order_type_enum') THEN
        CREATE TYPE work_order_type_enum AS ENUM ('Preventive', 'Corrective', 'Emergency');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'priority_level_enum') THEN
        CREATE TYPE priority_level_enum AS ENUM ('Low', 'Medium', 'High', 'Critical');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'work_order_status_enum') THEN
        CREATE TYPE work_order_status_enum AS ENUM ('Draft', 'Assigned', 'In Progress', 'Completed', 'Cancelled');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wo_number VARCHAR(50) NOT NULL,
    type work_order_type_enum NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    priority priority_level_enum NOT NULL,
    status work_order_status_enum NOT NULL DEFAULT 'Draft',
    reported_by BIGINT NOT NULL,
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_to_team VARCHAR(100) NOT NULL DEFAULT '',
    assigned_to_technician BIGINT NULL,
    assigned_at TIMESTAMPTZ NULL,
    scheduled_start TIMESTAMPTZ NULL,
    scheduled_end TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    estimated_duration_hours DECIMAL(6,2) NULL,
    actual_duration_hours DECIMAL(6,2) NULL,
    equipment_downtime_hours DECIMAL(6,2) NULL,
    estimated_cost DECIMAL(12,2) NULL,
    actual_labor_cost DECIMAL(12,2) NULL,
    actual_parts_cost DECIMAL(12,2) NULL,
    actual_total_cost DECIMAL(12,2) NULL,
    parts_used JSONB NOT NULL DEFAULT '[]'::JSONB,
    is_compliant BOOLEAN NOT NULL DEFAULT TRUE,
    compliance_notes TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]'::JSONB,
    technician_notes TEXT NOT NULL DEFAULT '',
    completion_report_url TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT valid_work_order_schedule CHECK (scheduled_end IS NULL OR scheduled_start IS NULL OR scheduled_end >= scheduled_start),
    CONSTRAINT valid_work_order_completion CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at),
    CONSTRAINT valid_work_order_downtime CHECK (equipment_downtime_hours IS NULL OR equipment_downtime_hours >= 0),
    CONSTRAINT valid_work_order_estimated_duration CHECK (estimated_duration_hours IS NULL OR estimated_duration_hours >= 0),
    CONSTRAINT valid_work_order_actual_duration CHECK (actual_duration_hours IS NULL OR actual_duration_hours >= 0),
    CONSTRAINT valid_work_order_estimated_cost CHECK (estimated_cost IS NULL OR estimated_cost >= 0),
    CONSTRAINT valid_work_order_actual_labor_cost CHECK (actual_labor_cost IS NULL OR actual_labor_cost >= 0),
    CONSTRAINT valid_work_order_actual_parts_cost CHECK (actual_parts_cost IS NULL OR actual_parts_cost >= 0),
    CONSTRAINT valid_work_order_actual_total_cost CHECK (actual_total_cost IS NULL OR actual_total_cost >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_work_orders_number_per_branch
    ON work_orders (business_id, branch_id, wo_number)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_orders_asset_id ON work_orders(asset_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
CREATE INDEX IF NOT EXISTS idx_work_orders_priority ON work_orders(priority);
CREATE INDEX IF NOT EXISTS idx_work_orders_assigned_technician ON work_orders(assigned_to_technician);
CREATE INDEX IF NOT EXISTS idx_work_orders_scheduled_start ON work_orders(scheduled_start);
CREATE INDEX IF NOT EXISTS idx_work_orders_reported_at ON work_orders(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_work_orders_business_branch_status ON work_orders(business_id, branch_id, status);
CREATE INDEX IF NOT EXISTS idx_work_orders_business_branch_priority ON work_orders(business_id, branch_id, priority);
CREATE INDEX IF NOT EXISTS idx_work_orders_business_branch_asset_status ON work_orders(business_id, branch_id, asset_id, status);
CREATE INDEX IF NOT EXISTS idx_work_orders_business_branch_type_status ON work_orders(business_id, branch_id, type, status);
CREATE INDEX IF NOT EXISTS idx_work_orders_active ON work_orders(business_id, branch_id, assigned_to_technician, scheduled_start)
    WHERE deleted_at IS NULL AND status NOT IN ('Completed', 'Cancelled');

CREATE TABLE IF NOT EXISTS maintenance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_order_id UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    technician_id BIGINT NOT NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NULL,
    parts_used JSONB NOT NULL DEFAULT '[]'::JSONB,
    images JSONB NOT NULL DEFAULT '[]'::JSONB,
    signature_url TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    CONSTRAINT valid_maintenance_log_duration CHECK (duration_minutes IS NULL OR duration_minutes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_maintenance_logs_work_order ON maintenance_logs(work_order_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_technician ON maintenance_logs(technician_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_created_at ON maintenance_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_business_branch ON maintenance_logs(business_id, branch_id);

CREATE TABLE IF NOT EXISTS preventive_maintenance_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    schedule_name VARCHAR(255) NOT NULL,
    frequency_type VARCHAR(50) NOT NULL
        CHECK (frequency_type IN ('Daily', 'Weekly', 'Monthly', 'Quarterly', 'Yearly', 'Custom')),
    frequency_value INTEGER NULL,
    frequency_unit VARCHAR(20) NULL,
    last_executed DATE NULL,
    next_due_date DATE NOT NULL,
    estimated_duration_hours DECIMAL(6,2) NULL,
    estimated_cost DECIMAL(12,2) NULL,
    checklist JSONB NOT NULL DEFAULT '[]'::JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT valid_pm_last_executed CHECK (last_executed IS NULL OR next_due_date >= last_executed),
    CONSTRAINT valid_pm_frequency_value CHECK (frequency_value IS NULL OR frequency_value > 0),
    CONSTRAINT valid_pm_estimated_duration CHECK (estimated_duration_hours IS NULL OR estimated_duration_hours >= 0),
    CONSTRAINT valid_pm_estimated_cost CHECK (estimated_cost IS NULL OR estimated_cost >= 0)
);

CREATE INDEX IF NOT EXISTS idx_pm_schedules_next_due
    ON preventive_maintenance_schedules(business_id, branch_id, next_due_date)
    WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_pm_schedules_asset ON preventive_maintenance_schedules(asset_id);
CREATE INDEX IF NOT EXISTS idx_pm_schedules_business_branch_asset
    ON preventive_maintenance_schedules(business_id, branch_id, asset_id);

COMMIT;
