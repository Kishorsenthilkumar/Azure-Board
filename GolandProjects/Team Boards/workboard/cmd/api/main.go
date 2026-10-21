package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"workboard/internal/database"
)

type application struct {
	db *sql.DB
}

func main() {

	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	db, err := database.Open()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("database connection established")

	server := http.Server{
		Addr:              *addr,
		Handler:           router(&application{db: db}),
		ReadHeaderTimeout: 5 * time.Second,
	}

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
