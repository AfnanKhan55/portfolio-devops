// Command server is the whole application: it serves the static portfolio
// website (index.html, css, images) and exposes /healthz + /version for
// Kubernetes. Everything downstream in this repo (Docker, Helm, CI, ArgoCD)
// exists to build, ship, and run this one small binary reliably.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AfnanKhan55/portfolio-devops/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	start := time.Now()
	mux := http.NewServeMux()

	// Serve the portfolio site's static files (index.html, images, etc.)
	// from ./web. In the container this directory is copied in at build time.
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// Kubernetes probes hit this to know the pod is alive and ready.
	mux.HandleFunc("/healthz", handlers.HealthCheck(start))

	// Lets you check which image/commit is actually deployed.
	mux.HandleFunc("/version", handlers.VersionHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlers.WithLogging(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run the server in a goroutine so we can listen for shutdown signals
	// on the main goroutine below (this is what makes rolling updates in
	// Kubernetes clean instead of dropping in-flight requests).
	go func() {
		log.Printf("portfolio server listening on :%s (version=%s)", port, handlers.Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for Ctrl+C locally, or SIGTERM from Kubernetes when it stops a pod.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
