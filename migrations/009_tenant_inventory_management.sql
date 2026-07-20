BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_number VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(100) NOT NULL DEFAULT '',
    compatible_assets JSONB NOT NULL DEFAULT '[]'::JSONB,
    manufacturer VARCHAR(255) NOT NULL DEFAULT '',
    supplier_id UUID NULL REFERENCES vendors(id) ON DELETE SET NULL,
    unit_of_measure VARCHAR(50) NOT NULL DEFAULT '',
    quantity_on_hand INTEGER NOT NULL DEFAULT 0,
    minimum_stock INTEGER NOT NULL DEFAULT 0,
    maximum_stock INTEGER NULL,
    reorder_quantity INTEGER NULL,
    unit_cost DECIMAL(12,2) NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT '',
    bin_location VARCHAR(100) NOT NULL DEFAULT '',
    last_restocked_at TIMESTAMPTZ NULL,
    last_restocked_by BIGINT NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    CONSTRAINT positive_quantity CHECK (quantity_on_hand >= 0),
    CONSTRAINT positive_minimum_stock CHECK (minimum_stock >= 0),
    CONSTRAINT positive_maximum_stock CHECK (maximum_stock IS NULL OR maximum_stock >= 0),
    CONSTRAINT positive_reorder_quantity CHECK (reorder_quantity IS NULL OR reorder_quantity >= 0),
    CONSTRAINT positive_unit_cost CHECK (unit_cost >= 0),
    CONSTRAINT valid_stock_range CHECK (minimum_stock <= maximum_stock OR maximum_stock IS NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_inventory_part_number_per_branch ON inventory_items (business_id, branch_id, LOWER(part_number));
CREATE INDEX IF NOT EXISTS idx_inventory_category ON inventory_items(category);
CREATE INDEX IF NOT EXISTS idx_inventory_supplier ON inventory_items(supplier_id);
CREATE INDEX IF NOT EXISTS idx_inventory_stock_status ON inventory_items(quantity_on_hand, minimum_stock);

CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_cost DECIMAL(12,2) NULL,
    reference_type VARCHAR(100) NOT NULL DEFAULT '',
    reference_id VARCHAR(100) NULL,
    work_order_id UUID NULL REFERENCES work_orders(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    performed_by BIGINT NOT NULL,
    performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    CONSTRAINT non_zero_transaction_quantity CHECK (quantity <> 0),
    CONSTRAINT positive_transaction_unit_cost CHECK (unit_cost IS NULL OR unit_cost >= 0)
);

CREATE INDEX IF NOT EXISTS idx_inventory_transactions_item ON inventory_transactions(item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_transactions_work_order ON inventory_transactions(work_order_id);
CREATE INDEX IF NOT EXISTS idx_inventory_transactions_performed_at ON inventory_transactions(performed_at DESC);

COMMIT;
