BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_code VARCHAR(50) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT '',
    cr_number VARCHAR(100) NULL,
    tax_number VARCHAR(100) NOT NULL DEFAULT '',
    contact_person VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    city VARCHAR(100) NOT NULL DEFAULT '',
    bank_account_details JSONB NOT NULL DEFAULT '{}'::JSONB,
    insurance_expiry DATE NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMPTZ NULL,
    approved_by BIGINT NULL,
    performance_rating DECIMAL(3,2) NULL,
    total_contracts_value DECIMAL(15,2) NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT valid_vendor_rating CHECK (performance_rating IS NULL OR (performance_rating >= 0 AND performance_rating <= 5)),
    CONSTRAINT valid_vendor_contract_value CHECK (total_contracts_value IS NULL OR total_contracts_value >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_vendors_code_per_branch
    ON vendors (business_id, branch_id, LOWER(vendor_code))
    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_vendors_cr_per_branch
    ON vendors (business_id, branch_id, LOWER(cr_number))
    WHERE deleted_at IS NULL AND cr_number IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_type ON vendors(type);
CREATE INDEX IF NOT EXISTS idx_vendors_approved ON vendors(approved);
CREATE INDEX IF NOT EXISTS idx_vendors_business_branch ON vendors(business_id, branch_id);

CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_number VARCHAR(50) NOT NULL,
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    scope_of_work TEXT NOT NULL DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    value_sar DECIMAL(15,2) NOT NULL,
    payment_terms TEXT NOT NULL DEFAULT '',
    sla_response_hours INTEGER NULL,
    sla_resolution_hours INTEGER NULL,
    penalties_clause TEXT NOT NULL DEFAULT '',
    documents JSONB NOT NULL DEFAULT '[]'::JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'Active',
    auto_renew BOOLEAN NOT NULL DEFAULT FALSE,
    renewal_notification_days INTEGER NOT NULL DEFAULT 90,
    signed_by_vendor_at TIMESTAMPTZ NULL,
    signed_by_facility_at TIMESTAMPTZ NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT valid_contract_dates CHECK (end_date >= start_date),
    CONSTRAINT valid_contract_value CHECK (value_sar >= 0),
    CONSTRAINT valid_contract_response_sla CHECK (sla_response_hours IS NULL OR sla_response_hours >= 0),
    CONSTRAINT valid_contract_resolution_sla CHECK (sla_resolution_hours IS NULL OR sla_resolution_hours >= 0),
    CONSTRAINT valid_contract_renewal_days CHECK (renewal_notification_days >= 0),
    CONSTRAINT valid_contract_status CHECK (status IN ('Active', 'Expired', 'Terminated', 'Draft'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_contracts_number_per_branch
    ON contracts (business_id, branch_id, LOWER(contract_number));
CREATE INDEX IF NOT EXISTS idx_contracts_vendor ON contracts(vendor_id);
CREATE INDEX IF NOT EXISTS idx_contracts_dates ON contracts(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);
CREATE INDEX IF NOT EXISTS idx_contracts_expiring ON contracts(end_date) WHERE status = 'Active';
CREATE INDEX IF NOT EXISTS idx_contracts_business_branch ON contracts(business_id, branch_id);

COMMIT;
