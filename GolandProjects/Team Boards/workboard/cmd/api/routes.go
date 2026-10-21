package main

import "net/http"

func router(app *application) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /healthz", app.HealthHandler)
	router.HandleFunc("GET /readyz", app.DatabaseHealthHandler)
	router.HandleFunc("POST /v1/projects", app.ProjectHandler)
	router.HandleFunc("GET /v1/projects", app.ListProjectsHandler)
	router.HandleFunc("GET /v1/projects/{id}", app.GetProjectHandler)
	router.HandleFunc("PUT /v1/projects/{id}", app.UpdateProjectHandler)
	router.HandleFunc("DELETE /v1/projects/{id}", app.DeleteProjectHandler)

	return router
}
