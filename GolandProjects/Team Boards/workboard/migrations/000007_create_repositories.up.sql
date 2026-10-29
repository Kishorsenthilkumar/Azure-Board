
CREATE TABLE repositories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    clone_url TEXT NOT NULL,
    default_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT repositories_provider_check
        CHECK (provider IN ('github', 'gitea')),

    CONSTRAINT repositories_project_name_provider_unique
        UNIQUE (project_id, provider, name),

    CONSTRAINT repositories_id_project_unique
    UNIQUE (id, project_id)
);

CREATE INDEX idx_repositories_project_id
    ON repositories(project_id);
