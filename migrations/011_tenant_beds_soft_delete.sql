BEGIN;

ALTER TABLE beds ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

DROP INDEX IF EXISTS uq_beds_room_number_per_branch;
CREATE UNIQUE INDEX IF NOT EXISTS uq_beds_room_number_per_branch
    ON beds (business_id, branch_id, room_id, bed_number)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_beds_deleted_at ON beds(deleted_at);

COMMIT;
