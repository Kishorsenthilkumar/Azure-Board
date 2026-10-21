CREATE TABLE work_items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    project_id BIGINT NOT NULL
        REFERENCES projects(id) ON DELETE CASCADE,

    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    work_item_type VARCHAR(30) NOT NULL
        CHECK (work_item_type IN ('Task', 'Bug', 'User Story')),

    task_type VARCHAR(20) NOT NULL
        CHECK (task_type IN ('Dev', 'Prod', 'Test')),

    status VARCHAR(20) NOT NULL DEFAULT 'New'
        CHECK (status IN ('New', 'Active', 'Resolved', 'Closed')),

    estimated_hours NUMERIC(6,2)
        CHECK (estimated_hours IS NULL OR estimated_hours >= 0),

    assigned_to BIGINT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_work_items_project_id
    ON work_items(project_id);

CREATE INDEX idx_work_items_status
    ON work_items(status);

CREATE INDEX idx_work_items_assigned_to
    ON work_items(assigned_to);
