DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'work_items_assigned_to_fkey'
          AND conrelid = 'work_items'::regclass
    ) THEN
        ALTER TABLE work_items
        ADD CONSTRAINT work_items_assigned_to_fkey
        FOREIGN KEY (assigned_to)
        REFERENCES users(id)
        ON DELETE SET NULL;
    END IF;
END $$;