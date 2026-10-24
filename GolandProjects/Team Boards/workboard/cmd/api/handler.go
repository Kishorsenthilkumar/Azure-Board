package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		res := map[string]string{
			"status": "service unavailable",
		}

		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println("failed to encode response:", err)
		}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		http.Error(w, "project name cannot be empty", http.StatusBadRequest)
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
		log.Printf("database error : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		log.Printf("database error : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var p project
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			log.Printf("database error : %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("database error : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		http.Error(w, "invalid project ID", http.StatusBadRequest)
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
		log.Printf("database error :%v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		http.Error(w, "project name cannot be empty", http.StatusBadRequest)
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
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM projects WHERE id = $1 RETURNING id;`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var deletedID int

	err = app.db.QueryRowContext(ctx, query, id).Scan(&deletedID)

	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) CreateWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	var input workItem

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	input.ProjectID = id

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		http.Error(w, "project title cannot be empty", http.StatusBadRequest)
		return
	}

	if input.Status == "" {
		input.Status = "New"
	}

	if input.EstimatedHours != nil && *input.EstimatedHours < 0 {
		http.Error(w, "estimated hours cannot be negative", http.StatusBadRequest)
		return
	}

	switch input.WorkItemType {
	case "Task", "Bug", "User Story":
		// Valid work-item type.
	default:
		http.Error(w, "invalid work-item type", http.StatusBadRequest)
		return
	}

	switch input.TaskType {
	case "Dev", "Prod", "Test":
		// Valid task type.
	default:
		http.Error(w, "invalid task type", http.StatusBadRequest)
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
		// Valid status.
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("database transaction error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Lock the project before calculating the next position.
	err = lockProject(ctx, tx, input.ProjectID)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("project lock error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var estimatedHours sql.NullFloat64

	query := `
    INSERT INTO work_items (
        project_id,
        title,
        description,
        work_item_type,
        task_type,
        status,
        estimated_hours,
        sort_order
    )
    VALUES (
        $1, $2, $3, $4, $5, $6, $7,
        COALESCE(
            (
                SELECT MAX(sort_order) + 1
                FROM work_items
                WHERE project_id = $1 AND status = $8
            ),
            1
        )
    )
    RETURNING
        id, project_id, title, description,
        work_item_type, task_type, status,
        estimated_hours, sort_order, created_at, updated_at;
`

	args := []interface{}{
		input.ProjectID,
		input.Title,
		input.Description,
		input.WorkItemType,
		input.TaskType,
		input.Status,
		input.EstimatedHours,
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
		&input.SortOrder,
		&input.CreatedAt,
		&input.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		// Foreign-key violation.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}

		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if estimatedHours.Valid {
		input.EstimatedHours = &estimatedHours.Float64
	} else {
		input.EstimatedHours = nil
	}

	// Commit before returning a successful response.
	if err := tx.Commit(); err != nil {
		log.Printf("database commit error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "invalid project ID", http.StatusBadRequest)
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
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
	}

	if workItemType != "" {
		switch workItemType {
		case "Task", "Bug", "User Story":
			// Valid
		default:
			http.Error(w, "invalid work item type", http.StatusBadRequest)
			return
		}
	}

	if taskType != "" {
		switch taskType {
		case "Dev", "Prod", "Test":
			// Valid
		default:
			http.Error(w, "invalid task type", http.StatusBadRequest)
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
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	query := `
    SELECT id, project_id, title, description,
           work_item_type, task_type, status,
           estimated_hours,sort_order, created_at, updated_at
    FROM work_items
    WHERE project_id = $1
      AND ($2 = '' OR status = $2)
      AND ($3 = '' OR work_item_type = $3)
      AND ($4 = '' OR task_type = $4)
    ORDER BY sort_order,id;
`

	rows, err := app.db.QueryContext(ctx, query, projectID, status, workItemType, taskType)
	if err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []workItem{}

	for rows.Next() {
		var item workItem
		var estimatedHours sql.NullFloat64

		err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&item.Title,
			&item.Description,
			&item.WorkItemType,
			&item.TaskType,
			&item.Status,
			&estimatedHours,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			log.Printf("database error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if estimatedHours.Valid {
			item.EstimatedHours = &estimatedHours.Float64
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "invalid work item ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT id, project_id, title, description,
               work_item_type, task_type, status,
               estimated_hours,sort_order, created_at, updated_at
        FROM work_items
        WHERE id = $1;
    `

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var item workItem
	var estimatedHours sql.NullFloat64

	err = app.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
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
		http.Error(w, "invalid work item ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		WorkItemType   string   `json:"work_item_type"`
		TaskType       string   `json:"task_type"`
		Status         string   `json:"status"`
		EstimatedHours *float64 `json:"estimated_hours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		http.Error(w, "title cannot be empty", http.StatusBadRequest)
		return
	}

	switch input.WorkItemType {
	case "Task", "Bug", "User Story":
	default:
		http.Error(w, "invalid work item type", http.StatusBadRequest)
		return
	}

	switch input.TaskType {
	case "Dev", "Prod", "Test":
	default:
		http.Error(w, "invalid task type", http.StatusBadRequest)
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if input.EstimatedHours != nil && *input.EstimatedHours < 0 {
		http.Error(w, "estimated hours cannot be negative", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("database transaction error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check whether the work item exists and move it if its status changes.
	err = moveWorkItemStatus(ctx, tx, id, input.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("status movement error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		    updated_at = NOW()
		WHERE id = $6
	`, input.Title, input.Description, input.WorkItemType,
		input.TaskType, input.EstimatedHours, id)

	if err != nil {
		log.Printf("work item update error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var item workItem
	var estimatedHours sql.NullFloat64

	err = tx.QueryRowContext(ctx, `
		SELECT id, project_id, title, description,
		       work_item_type, task_type, status,
		       estimated_hours, sort_order, created_at, updated_at
		FROM work_items
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("work item fetch error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		log.Printf("database commit error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "invalid work item ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("database transaction error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Step 2: Lock the project first.
	err = lockProject(ctx, tx, projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("project lock error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Step 4: Delete the work item.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM work_items
		WHERE id = $1
	`, id)

	if err != nil {
		log.Printf("work item deletion error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		log.Printf("sort order update error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Step 6: Commit both changes together.
	if err := tx.Commit(); err != nil {
		log.Printf("database commit error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) UpdateWorkItemStatusHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.Error(w, "invalid work item ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	switch input.Status {
	case "New", "Active", "Resolved", "Closed":
		// Valid status
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("database transaction error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Move the work item if its status has changed.
	err = moveWorkItemStatus(ctx, tx, id, input.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("status movement error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch the updated work item.
	var item workItem
	var estimatedHours sql.NullFloat64

	err = tx.QueryRowContext(ctx, `
		SELECT id, project_id, title, description,
		       work_item_type, task_type, status,
		       estimated_hours, sort_order, created_at, updated_at
		FROM work_items
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("work item fetch error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		log.Printf("database commit error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "invalid work item ID", http.StatusBadRequest)
		return
	}

	var input struct {
		SortOrder int64 `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if input.SortOrder < 1 {
		http.Error(w, "sort_order must be at least 1", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("database transaction error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = lockProject(ctx, tx, projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("project lock error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.QueryRowContext(ctx, `
		SELECT project_id, status, sort_order
		FROM work_items
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&projectID, &status, &oldOrder)

	if err == sql.ErrNoRows {
		http.Error(w, "work item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var maxOrder int64
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM work_items
		WHERE project_id = $1 AND status = $2
	`, projectID, status).Scan(&maxOrder)

	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
			log.Printf("reordering error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE work_items
			SET sort_order = $1,
			    updated_at = NOW()
			WHERE id = $2
		`, input.SortOrder, id)

		if err != nil {
			log.Printf("position update error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	var item workItem
	var estimatedHours sql.NullFloat64

	err = tx.QueryRowContext(ctx, `
		SELECT id, project_id, title, description,
		       work_item_type, task_type, status,
		       estimated_hours, sort_order, created_at, updated_at
		FROM work_items
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Description,
		&item.WorkItemType,
		&item.TaskType,
		&item.Status,
		&estimatedHours,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		log.Printf("database query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if estimatedHours.Valid {
		item.EstimatedHours = &estimatedHours.Float64
	}

	if err := tx.Commit(); err != nil {
		log.Printf("database commit error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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

func (app *application) CreateWorkItemComment(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.Error(w, "invalid work-item ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	var comment workItemDiscussion

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Content = strings.TrimSpace(input.Content)

	if input.Content == "" {
		http.Error(w, "comment cannot be empty", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var workItemID int64
	err = app.db.QueryRowContext(ctx,
		"SELECT id FROM work_items WHERE id = $1", id).Scan(&workItemID)

	if err == sql.ErrNoRows {
		http.Error(w, "workitem not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("database error %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO work_item_comments (work_item_id, content)
            VALUES ($1, $2)
            RETURNING id, work_item_id, content, created_at;`

	err = app.db.QueryRowContext(ctx, query, workItemID, input.Content).Scan(&comment.CommentID, &comment.WorkItemID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		log.Printf("database error :%v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
