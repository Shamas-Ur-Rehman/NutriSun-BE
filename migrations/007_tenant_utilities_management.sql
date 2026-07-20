BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'utility_type_enum') THEN
        CREATE TYPE utility_type_enum AS ENUM ('Electricity', 'Water', 'Gas', 'HVAC', 'Generator', 'Medical Oxygen', 'Nitrous Oxide', 'Compressed Air', 'Other');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS utility_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    utility_type utility_type_enum NOT NULL,
    meter_id VARCHAR(100) NOT NULL DEFAULT '',
    reading_value DECIMAL(12,2) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'Normal',
    alert_threshold DECIMAL(12,2) NULL,
    alert_triggered BOOLEAN NOT NULL DEFAULT FALSE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    recorded_by BIGINT NULL,
    notes TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    CONSTRAINT positive_utility_reading CHECK (reading_value >= 0),
    CONSTRAINT valid_utility_alert_threshold CHECK (alert_threshold IS NULL OR alert_threshold >= 0)
);

CREATE INDEX IF NOT EXISTS idx_utility_logs_type_time ON utility_logs(utility_type, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_utility_logs_meter_id ON utility_logs(meter_id);
CREATE INDEX IF NOT EXISTS idx_utility_logs_status ON utility_logs(status);
CREATE INDEX IF NOT EXISTS idx_utility_logs_business_branch_time ON utility_logs(business_id, branch_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS generator_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    generator_id VARCHAR(100) NOT NULL,
    test_date DATE NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    duration_minutes INTEGER NOT NULL,
    fuel_level_before DECIMAL(5,2) NULL,
    fuel_level_after DECIMAL(5,2) NULL,
    load_kw DECIMAL(10,2) NULL,
    passed BOOLEAN NOT NULL,
    issues_found TEXT NOT NULL DEFAULT '',
    next_test_due DATE NULL,
    tested_by BIGINT NOT NULL,
    report_url TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    CONSTRAINT positive_duration_minutes CHECK (duration_minutes > 0),
    CONSTRAINT valid_fuel_level_before CHECK (fuel_level_before IS NULL OR (fuel_level_before >= 0 AND fuel_level_before <= 100)),
    CONSTRAINT valid_fuel_level_after CHECK (fuel_level_after IS NULL OR (fuel_level_after >= 0 AND fuel_level_after <= 100)),
    CONSTRAINT valid_load_kw CHECK (load_kw IS NULL OR load_kw >= 0)
);

CREATE INDEX IF NOT EXISTS idx_generator_tests_test_date ON generator_tests(test_date DESC);
CREATE INDEX IF NOT EXISTS idx_generator_tests_generator ON generator_tests(generator_id);
CREATE INDEX IF NOT EXISTS idx_generator_tests_business_branch_date ON generator_tests(business_id, branch_id, test_date DESC);

COMMIT;
