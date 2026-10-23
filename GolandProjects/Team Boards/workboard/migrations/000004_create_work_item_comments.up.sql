CREATE TABLE work_item_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    work_item_id BIGINT NOT NULL
        REFERENCES work_items(id) ON DELETE CASCADE,
    content TEXT NOT NULL
        CHECK (length(trim(content)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_work_item_comments_work_item_id_created_at
    ON work_item_comments (work_item_id, created_at, id);