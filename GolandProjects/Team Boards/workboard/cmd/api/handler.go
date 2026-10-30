package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type workItem struct {
	ID             int64     `json:"id"`
	ProjectID      int64     `json:"project_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	WorkItemType   string    `json:"work_item_type"`
	TaskType       string    `json:"task_type"`
	Status         string    `json:"status"`
	AssignedTo     *int64    `json:"assigned_to"`
	AssignedName   *string   `json:"assigned_name"`
	EstimatedHours *float64  `json:"estimated_hours"`
	SortOrder      int64     `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type workItemDiscussion struct {
	CommentID  int64     `json:"comment_id"`
	WorkItemID int64     `json:"work_item_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type user struct {
	ID          int64     `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
}

type repository struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"project_id"`
	Provider      string    `json:"provider"`
	Name          string    `json:"name"`
	CloneURL      string    `json:"clone_url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (app *application) HealthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := map[string]string{
		"status": "ok",
	}

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("failed to encode response:", err)
	}

}

func (app *application) DatabaseHealthHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	err := app.db.PingContext(ctx)
	if err != nil {
		log.Printf("database connection failed %v", err)
		clientError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := map[string]string{
		"status": "ok",
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("failed to encode response:", err)
	}
}

func (app *application) ProjectHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	project := &project{}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		clientError(w, http.StatusBadRequest, "project name cannot be empty")
		return
	}

	query := `INSERT INTO projects (name, description)
              VALUES ($1, $2)
              RETURNING id, name, description, created_at, updated_at;`

	args := []interface{}{
		input.Name,
		input.Description,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, args...).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		log.Printf("json encoding :%v", err)
		return
	}

}

func (app *application) ListProjectsHandler(w http.ResponseWriter, r *http.Request) {

	projects := []project{}

	query := `SELECT id, name, description, created_at, updated_at
            FROM projects
            ORDER BY id;`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)

	if err != nil {
		serverError(w, r, err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var p project
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			serverError(w, r, err)
			return
		}
		projects = append(projects, p)
	}

	if err = rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(projects)
	if err != nil {
		log.Printf("json encoding :%v", err)
		return
	}
}

func (app *application) GetProjectHandler(w http.ResponseWriter, r *http.Request) {

	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	query := `SELECT id, name, description, created_at, updated_at
            FROM projects
            WHERE id = $1;`

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var p project

	err = app.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Printf("json encoding :%v", err)
		return
	}

}

func (app *application) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {

	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		clientError(w, http.StatusBadRequest, "project name cannot be empty")
		return
	}

	query := `
		UPDATE projects
		SET name = $1,
		    description = $2,
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at;
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var p project

	err = app.db.QueryRowContext(
		ctx,
		query,
		input.Name,
		input.Description,
		id,
	).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}

	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

func (app *application) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {

	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	query := `DELETE FROM projects WHERE id = $1 RETURNING id;`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var deletedID int

	err = app.db.QueryRowContext(ctx, query, id).Scan(&deletedID)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}

	if err != nil {
		serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) CreateWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var input workItem

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.ProjectID = id

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		clientError(w, http.StatusBadRequest, "project title cannot be empty")
		return
	}

	if input.Status == "" {
		input.Status = "New"
	}

	if input.EstimatedHours != nil && *input.EstimatedHours < 0 {
		clientError(w, http.StatusBadRequest, "estimated hours cannot be negative")
		return
	}

	switch input.WorkItemType {
	case "Task", "Bug", "User Story":
		// Valid work-item type.
	default:
		clientError(w, http.StatusBadRequest, "invalid work-item type")
		return
	}

	switch input.TaskType {
	case "Dev", "Prod", "Test":
		// Valid task type.
	default:
		clientError(w, http.StatusBadRequest, "invalid task type")
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
		// Valid status.
	default:
		clientError(w, http.StatusBadRequest, "invalid status")
		return
	}

	if input.AssignedTo != nil {
		var exists bool

		err := app.db.QueryRowContext(
			r.Context(),
			`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`,
			*input.AssignedTo,
		).Scan(&exists)

		if err != nil {
			serverError(w, r, err)
			return
		}

		if !exists {
			clientError(w, http.StatusBadRequest, "assigned user not found")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	// Lock the project before calculating the next position.
	err = lockProject(ctx, tx, input.ProjectID)
	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	var estimatedHours sql.NullFloat64
	var assignedTo sql.NullInt64

	query := `
    INSERT INTO work_items (
    project_id,
    title,
    description,
    work_item_type,
    task_type,
    status,
    estimated_hours,
    assigned_to,
    sort_order
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    COALESCE(
        (
            SELECT MAX(sort_order) + 1
            FROM work_items
            WHERE project_id = $1 AND status = $9
        ),
        1
    )
)
RETURNING
    id,
    project_id,
    title,
    description,
    work_item_type,
    task_type,
    status,
    estimated_hours,
    assigned_to,
    sort_order,
    created_at,
    updated_at;
`

	args := []interface{}{
		input.ProjectID,
		input.Title,
		input.Description,
		input.WorkItemType,
		input.TaskType,
		input.Status,
		input.EstimatedHours,
		input.AssignedTo,
		input.Status,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&input.ID,
		&input.ProjectID,
		&input.Title,
		&input.Description,
		&input.WorkItemType,
		&input.TaskType,
		&input.Status,
		&estimatedHours,
		&assignedTo,
		&input.SortOrder,
		&input.CreatedAt,
		&input.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		// Foreign-key violation.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			clientError(w, http.StatusNotFound, "project not found")
			return
		}

		serverError(w, r, err)
		return
	}

	if assignedTo.Valid {
		input.AssignedTo = &assignedTo.Int64
	} else {
		input.AssignedTo = nil
	}

	if estimatedHours.Valid {
		input.EstimatedHours = &estimatedHours.Float64
	} else {
		input.EstimatedHours = nil
	}

	if input.AssignedTo != nil {
		var assignedName string

		err = tx.QueryRowContext(
			ctx,
			`SELECT display_name FROM users WHERE id = $1`,
			*input.AssignedTo,
		).Scan(&assignedName)

		if err != nil {
			serverError(w, r, err)
			return
		}

		input.AssignedName = &assignedName
	} else {
		input.AssignedName = nil
	}

	// Commit before returning a successful response.
	if err := tx.Commit(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(input); err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}
func (app *application) ListWorkItemsHandler(w http.ResponseWriter, r *http.Request) {
	projectID, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	status := r.URL.Query().Get("status")
	workItemType := r.URL.Query().Get("work_item_type")
	taskType := r.URL.Query().Get("task_type")

	if status != "" {
		switch status {
		case "New", "Active", "Resolved", "Closed":
			// Valid status
		default:
			clientError(w, http.StatusBadRequest, "invalid status")
			return
		}
	}

	if workItemType != "" {
		switch workItemType {
		case "Task", "Bug", "User Story":
			// Valid
		default:
			clientError(w, http.StatusBadRequest, "invalid work item type")
			return
		}
	}

	if taskType != "" {
		switch taskType {
		case "Dev", "Prod", "Test":
			// Valid
		default:
			clientError(w, http.StatusBadRequest, "invalid task type")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check whether the project exists.
	var exists bool
	err = app.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`,
		projectID,
	).Scan(&exists)

	if err != nil {
		serverError(w, r, err)
		return
	}

	if !exists {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}

	query := `
     SELECT
    wi.id,
    wi.project_id,
    wi.title,
    wi.description,
    wi.work_item_type,
    wi.task_type,
    wi.status,
    wi.estimated_hours,
    wi.assigned_to,
    u.display_name AS assigned_name,
    wi.sort_order,
    wi.created_at,
    wi.updated_at
FROM work_items AS wi
LEFT JOIN users AS u ON u.id = wi.assigned_to
WHERE wi.project_id = $1
  AND ($2 = '' OR wi.status = $2)
  AND ($3 = '' OR wi.work_item_type = $3)
  AND ($4 = '' OR wi.task_type = $4)
ORDER BY wi.sort_order, wi.id;
`

	rows, err := app.db.QueryContext(ctx, query, projectID, status, workItemType, taskType)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer rows.Close()

	items := []workItem{}

	for rows.Next() {
		var item workItem
		var estimatedHours sql.NullFloat64
		var assignedTo sql.NullInt64
		var assignedName sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&item.Title,
			&item.Description,
			&item.WorkItemType,
			&item.TaskType,
			&item.Status,
			&estimatedHours,
			&assignedTo,
			&assignedName,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
			serverError(w, r, err)
			return
		}

		if assignedTo.Valid {
			item.AssignedTo = &assignedTo.Int64
		}

		if assignedName.Valid {
			item.AssignedName = &assignedName.String
		}

		if estimatedHours.Valid {
			item.EstimatedHours = &estimatedHours.Float64
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) GetWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	query := `SELECT
    wi.id,
    wi.project_id,
    wi.title,
    wi.description,
    wi.work_item_type,
    wi.task_type,
    wi.status,
    wi.estimated_hours,
    wi.assigned_to,
    u.display_name AS assigned_name,
    wi.sort_order,
    wi.created_at,
    wi.updated_at
FROM work_items AS wi
LEFT JOIN users AS u ON u.id = wi.assigned_to
WHERE wi.id = $1;`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var item workItem
	var estimatedHours sql.NullFloat64
	var assignedTo sql.NullInt64
	var assignedName sql.NullString

	err = app.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&assignedTo,
		&assignedName,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}

	if err != nil {
		serverError(w, r, err)
		return
	}

	if assignedTo.Valid {
		item.AssignedTo = &assignedTo.Int64
	}
	if assignedName.Valid {
		item.AssignedName = &assignedName.String
	}
	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) UpdateWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	var input struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		WorkItemType   string   `json:"work_item_type"`
		TaskType       string   `json:"task_type"`
		Status         string   `json:"status"`
		EstimatedHours *float64 `json:"estimated_hours"`
		AssignedTo     *int64   `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		clientError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	switch input.WorkItemType {
	case "Task", "Bug", "User Story":
	default:
		clientError(w, http.StatusBadRequest, "invalid work item type")
		return
	}

	switch input.TaskType {
	case "Dev", "Prod", "Test":
	default:
		clientError(w, http.StatusBadRequest, "invalid task type")
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
	default:
		clientError(w, http.StatusBadRequest, "invalid status")
		return
	}

	if input.EstimatedHours != nil && *input.EstimatedHours < 0 {
		clientError(w, http.StatusBadRequest, "estimated hours cannot be negative")
		return
	}

	if input.AssignedTo != nil {
		var exists bool

		err := app.db.QueryRowContext(
			r.Context(),
			`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`,
			*input.AssignedTo,
		).Scan(&exists)

		if err != nil {
			serverError(w, r, err)
			return
		}

		if !exists {
			clientError(w, http.StatusBadRequest, "assigned user not found")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	// Check whether the work item exists and move it if its status changes.
	err = moveWorkItemStatus(ctx, tx, id, input.Status)
	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	// Update the remaining fields without changing status or sort_order.
	_, err = tx.ExecContext(ctx, `
    UPDATE work_items
    SET title = $1,
        description = $2,
        work_item_type = $3,
        task_type = $4,
        estimated_hours = $5,
        assigned_to = $6,
        updated_at = NOW()
    WHERE id = $7
`,
		input.Title,
		input.Description,
		input.WorkItemType,
		input.TaskType,
		input.EstimatedHours,
		input.AssignedTo,
		id,
	)

	if err != nil {
		serverError(w, r, err)
		return
	}

	var item workItem
	var estimatedHours sql.NullFloat64
	var assignedTo sql.NullInt64
	var assignedName sql.NullString

	err = tx.QueryRowContext(ctx, `
    SELECT
        wi.id,
        wi.project_id,
        wi.title,
        wi.description,
        wi.work_item_type,
        wi.task_type,
        wi.status,
        wi.estimated_hours,
        wi.assigned_to,
        u.display_name AS assigned_name,
        wi.sort_order,
        wi.created_at,
        wi.updated_at
    FROM work_items AS wi
    LEFT JOIN users AS u ON u.id = wi.assigned_to
    WHERE wi.id = $1
`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&assignedTo,
		&assignedName,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	if assignedTo.Valid {
		item.AssignedTo = &assignedTo.Int64
	}

	if assignedName.Valid {
		item.AssignedName = &assignedName.String
	}

	if assignedTo.Valid {
		item.AssignedTo = &assignedTo.Int64
	}
	if assignedName.Valid {
		item.AssignedName = &assignedName.String
	}
	if assignedTo.Valid {
		item.AssignedTo = &assignedTo.Int64
	}
	if assignedName.Valid {
		item.AssignedName = &assignedName.String
	}
	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) DeleteWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	// Step 1: Find the project's ID without locking the work item.
	var projectID int64
	err = tx.QueryRowContext(ctx, `
		SELECT project_id
		FROM work_items
		WHERE id = $1
	`, id).Scan(&projectID)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	// Step 2: Lock the project first.
	err = lockProject(ctx, tx, projectID)
	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	// Step 3: Lock the work item and read its current column and position.
	var status string
	var oldOrder int64

	err = tx.QueryRowContext(ctx, `
		SELECT status, sort_order
		FROM work_items
		WHERE id = $1 AND project_id = $2
		FOR UPDATE
	`, id, projectID).Scan(&status, &oldOrder)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	// Step 4: Delete the work item.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM work_items
		WHERE id = $1
	`, id)

	if err != nil {
		serverError(w, r, err)
		return
	}

	// Step 5: Close the gap in the original column.
	_, err = tx.ExecContext(ctx, `
		UPDATE work_items
		SET sort_order = sort_order - 1
		WHERE project_id = $1
		  AND status = $2
		  AND sort_order > $3
	`, projectID, status, oldOrder)

	if err != nil {
		serverError(w, r, err)
		return
	}

	// Step 6: Commit both changes together.
	if err := tx.Commit(); err != nil {
		serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) UpdateWorkItemStatusHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
		// Valid status
	default:
		clientError(w, http.StatusBadRequest, "invalid status")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	// Move the work item if its status has changed.
	err = moveWorkItemStatus(ctx, tx, id, input.Status)
	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	// Fetch the updated work item.
	var item workItem
	var estimatedHours sql.NullFloat64

	var assignedTo sql.NullInt64
	var assignedName sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT wi.id, wi.project_id, wi.title, wi.description,
		       wi.work_item_type, wi.task_type, wi.status,
		       wi.estimated_hours, wi.assigned_to, u.display_name,
		       wi.sort_order, wi.created_at, wi.updated_at
		FROM work_items AS wi
		LEFT JOIN users AS u ON u.id = wi.assigned_to
		WHERE wi.id = $1
	`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&assignedTo,
		&assignedName,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) UpdateWorkItemPositionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	var input struct {
		SortOrder int64 `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if input.SortOrder < 1 {
		clientError(w, http.StatusBadRequest, "sort_order must be at least 1")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer tx.Rollback()

	var projectID int64
	var status string
	var oldOrder int64

	err = tx.QueryRowContext(ctx, `
    SELECT project_id
    FROM work_items
    WHERE id = $1
`, id).Scan(&projectID)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}

	if err != nil {
		serverError(w, r, err)
		return
	}

	err = lockProject(ctx, tx, projectID)
	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	err = tx.QueryRowContext(ctx, `
		SELECT project_id, status, sort_order
		FROM work_items
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&projectID, &status, &oldOrder)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	var maxOrder int64
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM work_items
		WHERE project_id = $1 AND status = $2
	`, projectID, status).Scan(&maxOrder)

	if err != nil {
		serverError(w, r, err)
		return
	}

	if input.SortOrder > maxOrder {
		input.SortOrder = maxOrder
	}

	if input.SortOrder != oldOrder {
		if input.SortOrder < oldOrder {
			_, err = tx.ExecContext(ctx, `
				UPDATE work_items
				SET sort_order = sort_order + 1
				WHERE project_id = $1
				  AND status = $2
				  AND id <> $3
				  AND sort_order >= $4
				  AND sort_order < $5
			`, projectID, status, id, input.SortOrder, oldOrder)
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE work_items
				SET sort_order = sort_order - 1
				WHERE project_id = $1
				  AND status = $2
				  AND id <> $3
				  AND sort_order > $4
				  AND sort_order <= $5
			`, projectID, status, id, oldOrder, input.SortOrder)
		}

		if err != nil {
			serverError(w, r, err)
			return
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE work_items
			SET sort_order = $1,
			    updated_at = NOW()
			WHERE id = $2
		`, input.SortOrder, id)

		if err != nil {
			serverError(w, r, err)
			return
		}
	}

	var item workItem
	var estimatedHours sql.NullFloat64

	var assignedTo sql.NullInt64
	var assignedName sql.NullString

	err = tx.QueryRowContext(ctx, `
		SELECT wi.id, wi.project_id, wi.title, wi.description,
		       wi.work_item_type, wi.task_type, wi.status,
		       wi.estimated_hours, wi.assigned_to, u.display_name,
		       wi.sort_order, wi.created_at, wi.updated_at
		FROM work_items AS wi
		LEFT JOIN users AS u ON u.id = wi.assigned_to
		WHERE wi.id = $1
	`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&assignedTo,
		&assignedName,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		serverError(w, r, err)
		return
	}

	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func moveWorkItemStatus(ctx context.Context, tx *sql.Tx, id int64, newStatus string) error {
	var projectID int64

	// Step 1: Read the project ID without locking the work item.
	err := tx.QueryRowContext(ctx, `
		SELECT project_id
		FROM work_items
		WHERE id = $1
	`, id).Scan(&projectID)
	if err != nil {
		return err
	}

	// Step 2: Lock the project before locking any work item.
	err = lockProject(ctx, tx, projectID)
	if err != nil {
		return err
	}

	var currentStatus string
	var currentOrder int64

	// Step 3: Lock the work item and read its current position.
	err = tx.QueryRowContext(ctx, `
		SELECT project_id, status, sort_order
		FROM work_items
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&projectID, &currentStatus, &currentOrder)
	if err != nil {
		return err
	}

	// Step 4: Preserve the current position if status is unchanged.
	if currentStatus == newStatus {
		return nil
	}

	// Step 5: Find the next position in the destination column.
	var destinationOrder int64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(sort_order), 0) + 1
		FROM work_items
		WHERE project_id = $1
		  AND status = $2
	`, projectID, newStatus).Scan(&destinationOrder)
	if err != nil {
		return err
	}

	// Step 6: Move the work item to the destination column.
	_, err = tx.ExecContext(ctx, `
		UPDATE work_items
		SET status = $1,
		    sort_order = $2,
		    updated_at = NOW()
		WHERE id = $3
	`, newStatus, destinationOrder, id)
	if err != nil {
		return err
	}

	// Step 7: Close the gap in the original column.
	_, err = tx.ExecContext(ctx, `
		UPDATE work_items
		SET sort_order = sort_order - 1
		WHERE project_id = $1
		  AND status = $2
		  AND sort_order > $3
	`, projectID, currentStatus, currentOrder)
	if err != nil {
		return err
	}

	return nil
}

func lockProject(ctx context.Context, tx *sql.Tx, projectID int64) error {

	err := tx.QueryRowContext(ctx, `
    SELECT id
    FROM projects
    WHERE id = $1
    FOR UPDATE
`, projectID).Scan(&projectID)

	return err

}

func (app *application) CreateWorkItemCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work-item ID")
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	var comment workItemDiscussion

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.Content = strings.TrimSpace(input.Content)

	if input.Content == "" {
		clientError(w, http.StatusBadRequest, "comment cannot be empty")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var workItemID int64
	err = app.db.QueryRowContext(ctx,
		"SELECT id FROM work_items WHERE id = $1", id).Scan(&workItemID)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "work item not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	query := `INSERT INTO work_item_comments (work_item_id, content)
            VALUES ($1, $2)
            RETURNING id, work_item_id, content, created_at;`

	err = app.db.QueryRowContext(ctx, query, workItemID, input.Content).Scan(&comment.CommentID, &comment.WorkItemID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

func (app *application) GetWorkItemCommentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid work-item ID")
		return
	}

	query := `SELECT id, work_item_id, content, created_at
			FROM work_item_comments
			WHERE work_item_id = $1
			ORDER BY created_at ASC, id ASC;  `

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	discussion := []workItemDiscussion{}

	q := `SELECT EXISTS (
			SELECT 1
			FROM work_items
			WHERE id = $1); `

	var found bool

	err = app.db.QueryRowContext(ctx, q, id).Scan(&found)
	if err != nil {
		serverError(w, r, err)
		return
	}

	if !found {
		clientError(w, http.StatusNotFound, "WorkItemID not found")
		return
	}

	rows, err := app.db.QueryContext(ctx, query, id)

	if err != nil {
		serverError(w, r, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var comment workItemDiscussion
		err = rows.Scan(&comment.CommentID, &comment.WorkItemID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			serverError(w, r, err)
			return
		}
		discussion = append(discussion, comment)
	}

	if err = rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(discussion)
	if err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

func (app *application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		clientError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Email = strings.TrimSpace(input.Email)

	if input.DisplayName == "" {
		clientError(w, http.StatusBadRequest, "display name cannot be empty")
		return
	}

	if input.Email == "" || !strings.Contains(input.Email, "@") {
		clientError(w, http.StatusBadRequest, "valid email is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var u user

	query := `
		INSERT INTO users (display_name, email)
		VALUES ($1, $2)
		RETURNING id, display_name, email, created_at;
	`

	err := app.db.QueryRowContext(
		ctx, query, input.DisplayName, input.Email,
	).Scan(&u.ID, &u.DisplayName, &u.Email, &u.CreatedAt)

	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users := []user{}

	query := `
		SELECT id, display_name, email, created_at
		FROM users
		ORDER BY display_name, id;
	`

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u user

		if err := rows.Scan(
			&u.ID,
			&u.DisplayName,
			&u.Email,
			&u.CreatedAt,
		); err != nil {
			serverError(w, r, err)
			return
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func (app *application) CreateRepositoryHandler(w http.ResponseWriter, r *http.Request) {

	id, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var input struct {
		Provider      string `json:"provider"`
		Name          string `json:"name"`
		CloneURL      string `json:"clone_url"`
		DefaultBranch string `json:"default_branch"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	input.Provider = strings.TrimSpace(input.Provider)
	input.Name = strings.TrimSpace(input.Name)
	input.CloneURL = strings.TrimSpace(input.CloneURL)
	input.DefaultBranch = strings.TrimSpace(input.DefaultBranch)

	switch input.Provider {
	case "github", "gitea":
		// Valid provider.
	default:
		clientError(w, http.StatusBadRequest, "invalid repository provider")
		return
	}

	if input.Name == "" {
		clientError(w, http.StatusBadRequest, "repository name cannot be empty")
		return
	}

	if input.CloneURL == "" {
		clientError(w, http.StatusBadRequest, "clone URL cannot be empty")
		return
	}

	if input.DefaultBranch == "" {
		input.DefaultBranch = "main"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var exists bool

	err = app.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
        SELECT 1 FROM projects WHERE id = $1
    )`,
		id,
	).Scan(&exists)

	if err != nil {
		serverError(w, r, err)
		return
	}

	if !exists {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}

	var repo repository

	query := `
    INSERT INTO repositories (
        project_id,
        provider,
        name,
        clone_url,
        default_branch
    )
    VALUES ($1, $2, $3, $4, $5)
    RETURNING
        id,
        project_id,
        provider,
        name,
        clone_url,
        default_branch,
        created_at,
        updated_at;
`

	err = app.db.QueryRowContext(
		ctx,
		query,
		id,
		input.Provider,
		input.Name,
		input.CloneURL,
		input.DefaultBranch,
	).Scan(
		&repo.ID,
		&repo.ProjectID,
		&repo.Provider,
		&repo.Name,
		&repo.CloneURL,
		&repo.DefaultBranch,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(repo); err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

func (app *application) ListRepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Read project ID.
	projectID, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Step 2: Create request context with timeout.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Step 3: Check whether the project exists.
	var exists bool

	err = app.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1 FROM projects WHERE id = $1
		)`,
		projectID,
	).Scan(&exists)

	if err != nil {
		serverError(w, r, err)
		return
	}

	if !exists {
		clientError(w, http.StatusNotFound, "project not found")
		return
	}

	// Step 4: Fetch repositories belonging to this project.
	query := `
		SELECT
			id,
			project_id,
			provider,
			name,
			clone_url,
			default_branch,
			created_at,
			updated_at
		FROM repositories
		WHERE project_id = $1
		ORDER BY id;
	`

	rows, err := app.db.QueryContext(ctx, query, projectID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer rows.Close()

	// Step 5: Scan repository records.
	repositories := []repository{}

	for rows.Next() {
		var repo repository

		err := rows.Scan(
			&repo.ID,
			&repo.ProjectID,
			&repo.Provider,
			&repo.Name,
			&repo.CloneURL,
			&repo.DefaultBranch,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		)

		if err != nil {
			serverError(w, r, err)
			return
		}

		repositories = append(repositories, repo)
	}

	// Step 6: Check iteration errors.
	if err := rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}

	// Step 7: Return response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(repositories); err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

func (app *application) GetRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	// Read project ID.
	projectID, err := readIDParam(r)
	if err != nil {
		clientError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Read repository ID.
	repositoryID, err := strconv.ParseInt(r.PathValue("repositoryID"), 10, 64)
	if err != nil || repositoryID < 1 {
		clientError(w, http.StatusBadRequest, "invalid repository ID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var repo repository

	query := `
		SELECT
			id,
			project_id,
			provider,
			name,
			clone_url,
			default_branch,
			created_at,
			updated_at
		FROM repositories
		WHERE id = $1 AND project_id = $2;
	`

	err = app.db.QueryRowContext(
		ctx,
		query,
		repositoryID,
		projectID,
	).Scan(
		&repo.ID,
		&repo.ProjectID,
		&repo.Provider,
		&repo.Name,
		&repo.CloneURL,
		&repo.DefaultBranch,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		clientError(w, http.StatusNotFound, "repository not found")
		return
	}

	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(repo); err != nil {
		log.Printf("JSON encoding error: %v", err)
		return
	}
}
