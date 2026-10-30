package main

import "net/http"

func router(app *application) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /healthz", app.HealthHandler)
	router.HandleFunc("GET /readyz", app.DatabaseHealthHandler)

	router.HandleFunc("POST /v1/projects", app.ProjectHandler)
	router.HandleFunc("GET /v1/projects", app.ListProjectsHandler)
	router.HandleFunc("GET /v1/projects/{id}", app.GetProjectHandler)
	router.HandleFunc("PUT /v1/projects/{id}", app.UpdateProjectHandler)
	router.HandleFunc("DELETE /v1/projects/{id}", app.DeleteProjectHandler)

	router.HandleFunc("POST /v1/projects/{id}/work-items", app.CreateWorkItemHandler)
	router.HandleFunc("GET /v1/projects/{id}/work-items", app.ListWorkItemsHandler)
	router.HandleFunc("GET /v1/work-items/{id}", app.GetWorkItemHandler)
	router.HandleFunc("PUT /v1/work-items/{id}", app.UpdateWorkItemHandler)
	router.HandleFunc("DELETE /v1/work-items/{id}", app.DeleteWorkItemHandler)

	router.HandleFunc("PATCH /v1/work-items/{id}/status", app.UpdateWorkItemStatusHandler)
	router.HandleFunc("PATCH /v1/work-items/{id}/position", app.UpdateWorkItemPositionHandler)

	router.HandleFunc("POST /v1/work-items/{id}/comments", app.CreateWorkItemCommentHandler)
	router.HandleFunc("GET /v1/work-items/{id}/comments", app.GetWorkItemCommentHandler)

	router.HandleFunc("POST /v1/users", app.CreateUserHandler)
	router.HandleFunc("GET /v1/users", app.ListUsersHandler)

	router.HandleFunc("POST /v1/projects/{id}/repositories", app.CreateRepositoryHandler)
	router.HandleFunc("GET /v1/projects/{id}/repositories", app.ListRepositoriesHandler)
	router.HandleFunc("GET /v1/projects/{id}/repositories/{repositoryID}", app.GetRepositoryHandler)

	return app.enableCORS(router)
}
