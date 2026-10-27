
ALTER TABLE work_items
ADD CONSTRAINT work_items_assigned_to_fkey
FOREIGN KEY (assigned_to)
REFERENCES users(id)
ON DELETE SET NULL;
