package pipeline

import "time"

type ClaimedPipelineRun struct {
	ID          int64
	PipelineID  int64
	TriggerType string
	Branch      string
	CommitSHA   string
	CreatedAt   time.Time
	StartedAt   time.Time
}
