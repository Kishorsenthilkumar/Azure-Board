
CREATE TABLE pipeline_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id BIGINT NOT NULL
        REFERENCES pipelines(id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    trigger_type VARCHAR(20) NOT NULL DEFAULT 'manual',

    branch VARCHAR(255) NOT NULL,
    commit_sha VARCHAR(64),

    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pipeline_runs_status_check
        CHECK (status IN (
            'queued',
            'running',
            'succeeded',
            'failed',
            'cancelled'
        )),

    CONSTRAINT pipeline_runs_trigger_type_check
        CHECK (trigger_type IN (
            'manual',
            'push',
            'pull_request'
        )),

    CONSTRAINT pipeline_runs_branch_not_empty
        CHECK (LENGTH(TRIM(branch)) > 0)
);

CREATE INDEX idx_pipeline_runs_pipeline_id_created_at
    ON pipeline_runs(pipeline_id, created_at DESC);

CREATE INDEX idx_pipeline_runs_status
    ON pipeline_runs(status);
