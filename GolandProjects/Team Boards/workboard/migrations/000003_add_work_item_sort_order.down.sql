DROP INDEX IF EXISTS idx_work_items_board_order;

ALTER TABLE work_items
DROP COLUMN IF EXISTS sort_order;