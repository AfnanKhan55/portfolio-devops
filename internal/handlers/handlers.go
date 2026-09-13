// Package handlers holds every HTTP handler and middleware the server uses.
// Keeping them here (instead of in main.go) makes them independently testable.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HealthResponse is what /healthz returns. Kubernetes liveness/readiness
// probes just care about the 200 status code, but a small JSON body makes
// manual debugging ("curl the pod") much nicer.
type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

// HealthCheck reports that the process is alive. This is the endpoint
// Kubernetes will call every few seconds to decide if the pod is healthy.
func HealthCheck(start time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status: "ok",
			Uptime: time.Since(start).String(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// Version is set from main via a build-time -ldflags value (see Makefile /
// Dockerfile). It lets you confirm which image is actually running,
// which matters a lot once GitHub Actions is tagging images automatically.
var Version = "dev"

// VersionHandler exposes the running build's version/commit.
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": Version})
}

// WithLogging wraps a handler so every request is logged. Simple stdout
// logging is enough here because Kubernetes/Docker capture stdout for you;
// nothing extra is needed to "get" logs off the container.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
