ALTER TABLE work_items
ADD COLUMN sort_order BIGINT NOT NULL DEFAULT 0;

WITH ordered_items AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY project_id, status
            ORDER BY id
        ) AS position
    FROM work_items
)
UPDATE work_items AS wi
SET sort_order = oi.position
FROM ordered_items AS oi
WHERE wi.id = oi.id;

CREATE INDEX idx_work_items_board_order
ON work_items (project_id, status, sort_order, id);