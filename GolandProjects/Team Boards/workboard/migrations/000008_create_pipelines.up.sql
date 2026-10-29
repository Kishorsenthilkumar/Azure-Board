
CREATE TABLE pipelines (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL,
    repository_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    config_path VARCHAR(1024) NOT NULL DEFAULT '.workboard/pipeline.yml',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pipelines_project_fk
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT pipelines_repository_project_fk
       FOREIGN KEY (repository_id, project_id)
       REFERENCES repositories(id, project_id)
    ON DELETE CASCADE,

    CONSTRAINT pipelines_name_not_empty
        CHECK (LENGTH(TRIM(name)) > 0),

    CONSTRAINT pipelines_config_path_not_empty
        CHECK (LENGTH(TRIM(config_path)) > 0),

    CONSTRAINT pipelines_project_name_unique
        UNIQUE (project_id, name)

);

CREATE INDEX idx_pipelines_repository_id
    ON pipelines(repository_id);
