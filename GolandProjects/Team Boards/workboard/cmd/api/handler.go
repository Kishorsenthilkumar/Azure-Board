package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

	idstr := strings.TrimSpace(r.PathValue("id"))

	if idstr == "" {
		http.Error(w, "project ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idstr)

	if err != nil {
		http.Error(w, "project ID must be a number", http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(w, "project ID should be greater than 0", http.StatusBadRequest)
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

	idStr := strings.TrimSpace(r.PathValue("id"))

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
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
	idStr := strings.TrimSpace(r.PathValue("id"))

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
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
