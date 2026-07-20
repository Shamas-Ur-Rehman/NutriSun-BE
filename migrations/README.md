# Database Migrations

## Order

1. Run `001_master_db.sql` against the master PostgreSQL database from `.env` `DB_NAME`.
2. Run `002_tenant_asset_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
3. Run `003_tenant_maintenance_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
4. Run `004_tenant_facility_operations.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
5. Run `005_tenant_space_bed_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
6. Run `006_tenant_vendor_contract_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
7. Run `007_tenant_utilities_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
8. Run `008_tenant_safety_compliance.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.
9. Run `009_tenant_inventory_management.sql` against each tenant PostgreSQL database stored in `businesses.fms_db`.

## What gets created

- Master DB:
  - `businesses`
  - `branches`
  - `subscriptions`
  - `roles`
  - `modules`
  - `permissions`
  - `users`
- Tenant DB:
  - `asset_categories`
  - `assets`
  - `asset_calibration_records`
  - `asset_maintenance_records`
  - `work_orders`
  - `maintenance_logs`
  - `preventive_maintenance_schedules`
  - `housekeeping_tasks`
  - `waste_records`
  - `rooms`
  - `beds`
  - `vendors`
  - `contracts`
  - `utility_logs`
  - `generator_tests`
  - `incidents`
  - `risk_assessments`
  - `inventory_items`
  - `inventory_transactions`

## Seed login

- Email: `admin@acmehospital.com`
- Password: `Admin@123`
- Business tenant DB value stored on business row: `acme_hospital_tenant`

## Notes

- `users.role_id` references `roles(id)` in the same master database.
- `fms_db` is still stored on the business record so the project can keep supporting separate business databases for future tenant-specific tables.
- The seeded admin user uses role `1`, which is created inside `001_master_db.sql`.
- `001_master_db.sql` now seeds the `asset` RBAC module and full admin permissions for it.
- `002_tenant_asset_management.sql` keeps `business_id` and `branch_id` on tenant asset tables for tenant-safe reporting and auditing.
- `003_tenant_maintenance_management.sql` follows the same tenant-safe pattern and keeps `business_id` and `branch_id` on every maintenance table.
- `004_tenant_facility_operations.sql` follows the same tenant-safe pattern for housekeeping and waste management tables.
- `005_tenant_space_bed_management.sql` adds tenant-scoped `rooms` and `beds` tables.
- `006_tenant_vendor_contract_management.sql` adds tenant-scoped `vendors` and `contracts` tables.
- `007_tenant_utilities_management.sql` adds tenant-scoped `utility_logs` and `generator_tests` tables.
- `008_tenant_safety_compliance.sql` adds tenant-scoped `incidents` and `risk_assessments` tables.
- `009_tenant_inventory_management.sql` adds tenant-scoped `inventory_items` and `inventory_transactions` tables.
- `011_tenant_beds_soft_delete.sql` adds soft-delete support for tenant beds.
- `facilities` tables do not exist in this repo yet, so `facility_id` remains a validated UUID reference instead of a foreign key for now.
- `housekeeping_tasks.room_id` and `waste_records.origin_room_id` remain UUID references for now because they were introduced before the room module and to keep migration ordering simple.
