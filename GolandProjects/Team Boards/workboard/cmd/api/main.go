package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {

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

func main() {

	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	router := http.NewServeMux()
	server := http.Server{
		Addr:              *addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	router.HandleFunc("GET /healthz", HealthHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {

		log.Printf("starting server on %s", server.Addr)

		serverErr <- server.ListenAndServe()

	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")

		shutdownctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdownctx)
		if err != nil {
			log.Printf("server shutdown failed %v", err)
		}

	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}

}
