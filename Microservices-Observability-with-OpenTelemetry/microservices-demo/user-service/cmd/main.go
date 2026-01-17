package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"user-service/internal/telemetry"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com"},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
}

func main() {
	ctx := context.Background()

	shutdown, err := telemetry.InitTracing(ctx)
	if err != nil {
		slog.Error("failed to init tracing", "error", err)
		os.Exit(1)
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("user-service")

	mux := http.NewServeMux()

	// /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "user-service",
		})
	})

	// /users
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "get_all_users")
		defer span.End()

		// Simulate DB delay
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		span.SetAttributes(
			attribute.Int("user.count", len(users)),
			attribute.String("operation.type", "read"),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		span.SetStatus(codes.Ok, "")
	})

	// /users/{id}
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/users/"):]
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		_, span := tracer.Start(r.Context(), "get_user_by_id")
		defer span.End()

		span.SetAttributes(
			attribute.Int("user.id", userID),
			attribute.String("operation.type", "read"),
		)

		// Simulate DB delay
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

		for _, u := range users {
			if u.ID == userID {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(u)
				span.SetStatus(codes.Ok, "")
				return
			}
		}

		err = http.ErrNoLocation
		span.RecordError(err)
		span.SetStatus(codes.Error, "User not found")
		http.Error(w, "User not found", http.StatusNotFound)
	})

	slog.Info("Listening on :8888")
	server := &http.Server{
		Addr:    ":8888",
		Handler: otelhttp.NewHandler(mux, "http-server"),
	}
	log.Fatal(server.ListenAndServe())
}
