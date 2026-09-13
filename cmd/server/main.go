package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AfnanKhan55/portfolio-devops/internal/db"
	"github.com/AfnanKhan55/portfolio-devops/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	start := time.Now()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://portfolio:portfolio@postgres:5432/portfolio?sslmode=disable"
	}

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	if err := db.EnsureSchema(context.Background(), conn); err != nil {
		log.Fatalf("db schema: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.HandleFunc("/healthz", handlers.HealthCheck(start))
	mux.HandleFunc("/version", handlers.VersionHandler)
	mux.HandleFunc("/visits", handlers.VisitsHandler(conn))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlers.WithLogging(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("server stopped")
}
