package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"order-service/internal/telemetry"
)

type Item struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price,omitempty"`
}

type Order struct {
	ID     int     `json:"id"`
	UserID int     `json:"user_id"`
	Items  []Item  `json:"items"`
	Total  float64 `json:"total"`
	Status string  `json:"status"`
}

var orders = []Order{
	{ID: 1, UserID: 1, Items: []Item{{ProductID: 101, Quantity: 2}}, Total: 99.98, Status: "completed"},
	{ID: 2, UserID: 2, Items: []Item{{ProductID: 102, Quantity: 1}}, Total: 49.99, Status: "pending"},
}

func getUserServiceURL() string {
	if url := os.Getenv("USER_SERVICE_URL"); url != "" {
		return url
	}
	return "http://localhost:8888"
}

func main() {
	ctx := context.Background()

	shutdown, err := telemetry.InitTracing(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("order-service")

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   5 * time.Second,
	}

	mux := http.NewServeMux()

	// /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "order-service",
		})
	})

	// GET /orders
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createOrder(w, r, tracer, &client)
			return
		}

		_, span := tracer.Start(r.Context(), "get_all_orders")
		defer span.End()

		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		span.SetAttributes(
			attribute.Int("order.count", len(orders)),
			attribute.String("operation.type", "read"),
		)

		json.NewEncoder(w).Encode(orders)
		span.SetStatus(codes.Ok, "")
	})

	// GET /orders/{id}
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/orders/"):]
		orderID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid order id", http.StatusBadRequest)
			return
		}

		_, span := tracer.Start(r.Context(), "get_order_by_id")
		defer span.End()

		span.SetAttributes(
			attribute.Int("order.id", orderID),
			attribute.String("operation.type", "read"),
		)

		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

		for _, o := range orders {
			if o.ID == orderID {
				json.NewEncoder(w).Encode(o)
				span.SetStatus(codes.Ok, "")
				return
			}
		}

		span.SetStatus(codes.Error, "Order not found")
		http.Error(w, "Order not found", http.StatusNotFound)
	})

	server := &http.Server{
		Addr:    ":3002",
		Handler: otelhttp.NewHandler(mux, "http-server"),
	}

	log.Println("Order service listening on port 3002")
	log.Fatal(server.ListenAndServe())
}

func createOrder(
	w http.ResponseWriter,
	r *http.Request,
	tracer trace.Tracer,
	client *http.Client,
) {
	ctx, span := tracer.Start(r.Context(), "create_order")
	defer span.End()

	var payload struct {
		UserID int    `json:"user_id"`
		Items  []Item `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int("order.user_id", payload.UserID),
		attribute.Int("order.items_count", len(payload.Items)),
		attribute.String("operation.type", "write"),
	)

	// Verify user
	func() {
		_, userSpan := tracer.Start(ctx, "verify_user")
		defer userSpan.End()

		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			getUserServiceURL()+"/users/"+strconv.Itoa(payload.UserID),
			nil,
		)

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			userSpan.RecordError(err)
			http.Error(w, "Invalid user", http.StatusBadRequest)
			return
		}
	}()

	// TODO: Check inventory

	total := 0.0
	for _, item := range payload.Items {
		price := item.Price
		if price == 0 {
			price = 25.99
		}
		total += price * float64(item.Quantity)
	}

	newOrder := Order{
		ID:     len(orders) + 1,
		UserID: payload.UserID,
		Items:  payload.Items,
		Total:  total,
		Status: "pending",
	}

	orders = append(orders, newOrder)

	span.SetAttributes(
		attribute.Int("order.id", newOrder.ID),
		attribute.Float64("order.total", newOrder.Total),
	)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}
