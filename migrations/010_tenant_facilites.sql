BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL DEFAULT '',

    code VARCHAR(50) NOT NULL,
    type VARCHAR(100) NOT NULL DEFAULT '',

    building VARCHAR(100) NOT NULL DEFAULT '',
    floor VARCHAR(50) NOT NULL DEFAULT '',

    capacity INTEGER NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'Active'
        CHECK (status IN ('Active', 'Inactive', 'Under Maintenance')),

    description TEXT NOT NULL DEFAULT '',

    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,

    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_facilities_code_per_branch
    ON facilities (business_id, branch_id, LOWER(code))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_facilities_name_per_branch
    ON facilities (business_id, branch_id, LOWER(name_en))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_facilities_business_branch
    ON facilities (business_id, branch_id);

CREATE INDEX IF NOT EXISTS idx_facilities_status
    ON facilities(status);

CREATE INDEX IF NOT EXISTS idx_facilities_created_at
    ON facilities(created_at DESC);

INSERT INTO facilities (
    name_en,
    name_ar,
    code,
    type,
    building,
    floor,
    capacity,
    status,
    description,
    business_id,
    branch_id,
    created_by,
    updated_by
)
VALUES
(
    'ICU Unit A',
    'وحدة العناية المركزة',
    'ICU-A1',
    'ICU',
    'Block A',
    '3',
    10,
    'Active',
    'Critical care unit',
    1,
    1,
    1,
    1
)
ON CONFLICT DO NOTHING;

COMMIT;