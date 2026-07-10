-- Migration 002: Extend notifications table with priority and metadata columns
-- Replaces old 'data' JSONB with 'metadata' and adds 'priority' field.
-- Safe to run multiple times (IF NOT EXISTS / idempotent pattern).

-- Add priority column (if not present)
ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS priority VARCHAR(20) DEFAULT 'low';

-- Rename data → metadata (only if data column still exists and metadata does not)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'notifications' AND column_name = 'data'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'notifications' AND column_name = 'metadata'
    ) THEN
        ALTER TABLE notifications RENAME COLUMN data TO metadata;
    END IF;
END
$$;

-- Add metadata column if neither data nor metadata exist
ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Composite index for fast unread count queries
CREATE INDEX IF NOT EXISTS idx_notifs_user_read_created
    ON notifications(user_id, is_read, created_at DESC);

-- Index for type-filtered queries
CREATE INDEX IF NOT EXISTS idx_notifs_user_type
    ON notifications(user_id, type);
