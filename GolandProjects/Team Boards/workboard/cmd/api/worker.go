package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

    "workboard/internal/pipeline"
)

func (app *application) claimPipelineRun(ctx context.Context) (*pipeline.ClaimedPipelineRun, error) {
	query := `
        WITH candidate AS (
            SELECT id
            FROM pipeline_runs
            WHERE status = 'queued'
            ORDER BY created_at, id
            FOR UPDATE SKIP LOCKED
            LIMIT 1
        )
        UPDATE pipeline_runs AS pr
        SET
            status = 'running',
            started_at = NOW()
        FROM candidate
        WHERE pr.id = candidate.id
        RETURNING
            pr.id,
            pr.pipeline_id,
            pr.trigger_type,
            pr.branch,
            pr.commit_sha,
            pr.created_at,
            pr.started_at
    `

    var run pipeline.ClaimedPipelineRun

	err := app.db.QueryRowContext(ctx, query).Scan(
		&run.ID,
		&run.PipelineID,
		&run.TriggerType,
		&run.Branch,
		&run.CommitSHA,
		&run.CreatedAt,
		&run.StartedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("claim pipeline run: %w", err)
	}

	return &run, nil
}
