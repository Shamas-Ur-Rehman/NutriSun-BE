BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'incident_severity_enum') THEN
        CREATE TYPE incident_severity_enum AS ENUM ('Minor', 'Moderate', 'Major', 'Critical');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_number VARCHAR(50) NOT NULL,
    type VARCHAR(100) NOT NULL,
    severity incident_severity_enum NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT '',
    room_id UUID NULL REFERENCES rooms(id) ON DELETE SET NULL,
    reported_by BIGINT NOT NULL,
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    description TEXT NOT NULL,
    immediate_action TEXT NOT NULL DEFAULT '',
    root_cause TEXT NOT NULL DEFAULT '',
    corrective_actions TEXT NOT NULL DEFAULT '',
    investigation_complete BOOLEAN NOT NULL DEFAULT FALSE,
    investigation_completed_at TIMESTAMPTZ NULL,
    investigated_by BIGINT NULL,
    reported_to_cbahi BOOLEAN NOT NULL DEFAULT FALSE,
    reported_to_cbahi_at TIMESTAMPTZ NULL,
    attachments JSONB NOT NULL DEFAULT '[]'::JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'Open',
    closed_at TIMESTAMPTZ NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT valid_incident_investigation_time CHECK (investigation_completed_at IS NULL OR investigation_completed_at >= reported_at),
    CONSTRAINT valid_incident_cbahi_time CHECK (reported_to_cbahi_at IS NULL OR reported_to_cbahi_at >= reported_at),
    CONSTRAINT valid_incident_closed_time CHECK (closed_at IS NULL OR closed_at >= reported_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_incidents_number_per_branch ON incidents (business_id, branch_id, incident_number);
CREATE INDEX IF NOT EXISTS idx_incidents_reported_at ON incidents(reported_at DESC);
CREATE INDEX IF NOT EXISTS idx_incidents_type ON incidents(type);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_cbahi ON incidents(reported_to_cbahi);

CREATE TABLE IF NOT EXISTS risk_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    risk_number VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT '',
    likelihood INTEGER NOT NULL CHECK (likelihood BETWEEN 1 AND 5),
    impact INTEGER NOT NULL CHECK (impact BETWEEN 1 AND 5),
    risk_score INTEGER GENERATED ALWAYS AS (likelihood * impact) STORED,
    mitigation_plan TEXT NOT NULL DEFAULT '',
    responsible_person BIGINT NULL,
    target_completion_date DATE NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Open',
    reviewed_at TIMESTAMPTZ NULL,
    reviewed_by BIGINT NULL,
    closed_at TIMESTAMPTZ NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT valid_risk_review_time CHECK (reviewed_at IS NULL OR reviewed_at >= created_at),
    CONSTRAINT valid_risk_closed_time CHECK (closed_at IS NULL OR closed_at >= created_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_risk_assessments_number_per_branch ON risk_assessments (business_id, branch_id, risk_number);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_score ON risk_assessments(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_status ON risk_assessments(status);

COMMIT;
