BEGIN;

CREATE TABLE IF NOT EXISTS businesses (
    id BIGSERIAL PRIMARY KEY,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    business_type VARCHAR(30) NOT NULL CHECK (business_type IN ('food_delivery', 'meal_subscription')),
    address TEXT NOT NULL DEFAULT '',
    contact_info TEXT NOT NULL DEFAULT '',
    vat_no VARCHAR(100) NOT NULL,
    cr_no VARCHAR(100) NOT NULL,
    city VARCHAR(120) NOT NULL DEFAULT '',
    vat_registration_date VARCHAR(50) NOT NULL DEFAULT '',
    logo TEXT NOT NULL DEFAULT '',
    signature TEXT NOT NULL DEFAULT '',
    stamp TEXT NOT NULL DEFAULT '',
    rcm_email VARCHAR(255) NOT NULL DEFAULT '',
    rcm_password TEXT NOT NULL DEFAULT '',
    latitude VARCHAR(50) NOT NULL DEFAULT '',
    longitude VARCHAR(50) NOT NULL DEFAULT '',
    license_no VARCHAR(150) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL DEFAULT '',
    registration_no VARCHAR(150) NOT NULL DEFAULT '',
    privacy_policy TEXT NOT NULL DEFAULT '',
    db VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_businesses_name_en ON businesses (LOWER(name_en));
CREATE UNIQUE INDEX IF NOT EXISTS uq_businesses_name_ar ON businesses (LOWER(name_ar));
CREATE UNIQUE INDEX IF NOT EXISTS uq_businesses_vat_no ON businesses (vat_no);
CREATE UNIQUE INDEX IF NOT EXISTS uq_businesses_cr_no ON businesses (cr_no);
CREATE UNIQUE INDEX IF NOT EXISTS uq_businesses_db ON businesses (db);

CREATE TABLE IF NOT EXISTS branches (
    id BIGSERIAL PRIMARY KEY,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_branches_business_id ON branches (business_id);
CREATE INDEX IF NOT EXISTS idx_branches_deleted_at ON branches (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_branches_business_name_en ON branches (business_id, LOWER(name_en));
CREATE UNIQUE INDEX IF NOT EXISTS uq_branches_business_name_ar ON branches (business_id, LOWER(name_ar));

CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    api_key VARCHAR(255) NOT NULL,
    secret_key VARCHAR(255) NOT NULL DEFAULT '',
    radiology_url TEXT NOT NULL DEFAULT '',
    radiology_api_key VARCHAR(255) NOT NULL DEFAULT '',
    lab_url TEXT NOT NULL DEFAULT '',
    lab_api_key VARCHAR(255) NOT NULL DEFAULT '',
    his_url TEXT NOT NULL DEFAULT '',
    his_api_key VARCHAR(255) NOT NULL DEFAULT '',
    branch_id BIGINT REFERENCES branches(id) ON DELETE SET NULL,
    expiry_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'expired', 'suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_subscriptions_api_key ON subscriptions (api_key);
CREATE INDEX IF NOT EXISTS idx_subscriptions_business_id ON subscriptions (business_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_branch_id ON subscriptions (branch_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions (deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_name ON roles (LOWER(name));
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE IF NOT EXISTS modules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_modules_name ON modules (LOWER(name));
CREATE INDEX IF NOT EXISTS idx_modules_deleted_at ON modules (deleted_at);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    module_id BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL CHECK (action IN ('get', 'create', 'update', 'delete')),
    business_id BIGINT NOT NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_role_module_action
    ON permissions (role_id, module_id, action);
CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON permissions (deleted_at);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    employee_id VARCHAR(100) NOT NULL,
    contact VARCHAR(100) NOT NULL DEFAULT '',
    date_of_birth DATE NULL,
    gender VARCHAR(10) NOT NULL DEFAULT 'other' CHECK (gender IN ('male', 'female', 'other')),
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_staff BOOLEAN NOT NULL DEFAULT FALSE,
    service VARCHAR(20) NOT NULL DEFAULT 'nutrisun' CHECK (service IN ('nutrisun')),
    address TEXT NOT NULL DEFAULT '',
    nationality VARCHAR(120) NOT NULL DEFAULT '',
    document_id VARCHAR(150) NOT NULL DEFAULT '',
    license VARCHAR(150) NOT NULL DEFAULT '',
    fcm_token TEXT NOT NULL DEFAULT '',
    device_id VARCHAR(255) NOT NULL DEFAULT '',
    device_type VARCHAR(100) NOT NULL DEFAULT '',
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE RESTRICT,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email ON users (LOWER(email));
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_employee_id ON users (employee_id);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);
CREATE INDEX IF NOT EXISTS idx_users_business_id ON users (business_id);
CREATE INDEX IF NOT EXISTS idx_users_branch_id ON users (branch_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

INSERT INTO businesses (
    id, name_en, name_ar, business_type, address, contact_info, vat_no, cr_no, city,
    vat_registration_date, logo, signature, stamp, rcm_email, rcm_password, latitude,
    longitude, license_no, email, registration_no, privacy_policy, db, created_at, updated_at
)
VALUES (
    1, 'NutriSun', 'NutriSun', 'food_delivery', 'Lahore, Pakistan', '+92-300-0000000',
    'VAT-NUTRISUN-001', 'CR-NUTRISUN-001', 'Lahore', '2026-07-20', '', '', '',
    '', '', '', '', 'LIC-NUTRISUN-001', 'admin@nutrisun.com',
    'REG-NUTRISUN-001', '', 'nutrisun_tenant', NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO branches (
    id, name_en, name_ar, business_id, created_at, updated_at
)
VALUES (
    1, 'Main Branch', 'Main Branch', 1, NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO subscriptions (
    id, business_id, api_key, secret_key, branch_id, expiry_date, status, created_at, updated_at
)
VALUES (
    1, 1, 'nutrisun-master-api-key', 'nutrisun-master-secret', 1, NOW() + INTERVAL '365 days', 'active', NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO roles (id, name, business_id, branch_id, created_at, updated_at)
VALUES (1, 'admin', 1, 1, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO modules (id, name, display_name, description, business_id, branch_id, created_at, updated_at)
VALUES
    (1, 'role', 'Role Management', 'Manage roles, permissions, and modules', 1, 1, NOW(), NOW()),
    (2, 'user', 'User Management', 'Manage tenant users from master identity records', 1, 1, NOW(), NOW()),
    (3, 'customer', 'Customer Management', 'Manage NutriSun customers', 1, 1, NOW(), NOW()),
    (4, 'customer_address', 'Customer Address Management', 'Manage NutriSun customer addresses', 1, 1, NOW(), NOW()),
    (5, 'subscription_plan', 'Subscription Plan Management', 'Manage NutriSun subscription plans', 1, 1, NOW(), NOW()),
    (6, 'menu', 'Menu Management', 'Manage NutriSun monthly menus', 1, 1, NOW(), NOW()),
    (7, 'customer_subscription', 'Customer Subscription Management', 'Manage NutriSun customer subscriptions', 1, 1, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (role_id, module_id, action, business_id, branch_id, created_at, updated_at)
VALUES
    (1, 1, 'get', 1, 1, NOW(), NOW()),
    (1, 1, 'create', 1, 1, NOW(), NOW()),
    (1, 1, 'update', 1, 1, NOW(), NOW()),
    (1, 1, 'delete', 1, 1, NOW(), NOW()),
    (1, 2, 'get', 1, 1, NOW(), NOW()),
    (1, 2, 'create', 1, 1, NOW(), NOW()),
    (1, 2, 'update', 1, 1, NOW(), NOW()),
    (1, 2, 'delete', 1, 1, NOW(), NOW()),
    (1, 3, 'get', 1, 1, NOW(), NOW()),
    (1, 3, 'create', 1, 1, NOW(), NOW()),
    (1, 3, 'update', 1, 1, NOW(), NOW()),
    (1, 3, 'delete', 1, 1, NOW(), NOW()),
    (1, 4, 'get', 1, 1, NOW(), NOW()),
    (1, 4, 'create', 1, 1, NOW(), NOW()),
    (1, 4, 'update', 1, 1, NOW(), NOW()),
    (1, 4, 'delete', 1, 1, NOW(), NOW()),
    (1, 5, 'get', 1, 1, NOW(), NOW()),
    (1, 5, 'create', 1, 1, NOW(), NOW()),
    (1, 5, 'update', 1, 1, NOW(), NOW()),
    (1, 5, 'delete', 1, 1, NOW(), NOW()),
    (1, 6, 'get', 1, 1, NOW(), NOW()),
    (1, 6, 'create', 1, 1, NOW(), NOW()),
    (1, 6, 'update', 1, 1, NOW(), NOW()),
    (1, 6, 'delete', 1, 1, NOW(), NOW()),
    (1, 7, 'get', 1, 1, NOW(), NOW()),
    (1, 7, 'create', 1, 1, NOW(), NOW()),
    (1, 7, 'update', 1, 1, NOW(), NOW()),
    (1, 7, 'delete', 1, 1, NOW(), NOW()),
    (1, 3, 'get', 1, 1, NOW(), NOW()),
    (1, 3, 'create', 1, 1, NOW(), NOW()),
    (1, 3, 'update', 1, 1, NOW(), NOW()),
    (1, 3, 'delete', 1, 1, NOW(), NOW()),
    (1, 4, 'get', 1, 1, NOW(), NOW()),
    (1, 4, 'create', 1, 1, NOW(), NOW()),
    (1, 4, 'update', 1, 1, NOW(), NOW()),
    (1, 4, 'delete', 1, 1, NOW(), NOW()),
    (1, 5, 'get', 1, 1, NOW(), NOW()),
    (1, 5, 'create', 1, 1, NOW(), NOW()),
    (1, 5, 'update', 1, 1, NOW(), NOW()),
    (1, 5, 'delete', 1, 1, NOW(), NOW()),
    (1, 6, 'get', 1, 1, NOW(), NOW()),
    (1, 6, 'create', 1, 1, NOW(), NOW()),
    (1, 6, 'update', 1, 1, NOW(), NOW()),
    (1, 6, 'delete', 1, 1, NOW(), NOW()),
    (1, 7, 'get', 1, 1, NOW(), NOW()),
    (1, 7, 'create', 1, 1, NOW(), NOW()),
    (1, 7, 'update', 1, 1, NOW(), NOW()),
    (1, 7, 'delete', 1, 1, NOW(), NOW())
ON CONFLICT (role_id, module_id, action) DO NOTHING;

INSERT INTO users (
    id, full_name, email, password, employee_id, contact, gender, role_id, is_active, is_staff,
    service, address, nationality, document_id, license, fcm_token, device_id, device_type,
    business_id, branch_id, created_at, updated_at
)
VALUES (
    1, 'NutriSun Admin', 'admin@nutrisun.com',
    '$2a$10$qvSQkI5lY5JEyxkS.jqKQu4f/xgFPaOC65P/Jsy4Ohsx8bYTaFgOC',
    'NSEMP0001', '+92-300-0000000', 'male', 1, TRUE, TRUE, 'nutrisun',
    'Lahore, Pakistan', 'Pakistani', 'DOC-ADMIN-001', '', '', '', '',
    1, 1, NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

SELECT setval('businesses_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM businesses), 1), true);
SELECT setval('branches_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM branches), 1), true);
SELECT setval('subscriptions_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM subscriptions), 1), true);
SELECT setval('roles_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM roles), 1), true);
SELECT setval('modules_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM modules), 1), true);
SELECT setval('permissions_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM permissions), 1), true);
SELECT setval('users_id_seq', GREATEST((SELECT COALESCE(MAX(id), 1) FROM users), 1), true);

COMMIT;
