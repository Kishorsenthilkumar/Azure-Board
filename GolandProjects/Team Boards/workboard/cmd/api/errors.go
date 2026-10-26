package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Request-method:%s ,url:%s, Error :%v", r.Method, r.URL.Path, err)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusInternalServerError)
	err = json.NewEncoder(w).Encode(map[string]string{
		"error": "internal server error",
	})
	if err != nil {
		log.Printf("json encoding :%v", err)
		return
	}
}

func clientError(w http.ResponseWriter, status int, message string) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
	if err != nil {
		log.Printf("json encoding :%v", err)
		return
	}

}

func notFoundError(w http.ResponseWriter) {
	clientError(w, http.StatusNotFound, "resource not found")
}
