package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"inventory-service/internal/telemetry"
)

type InventoryItem struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Stock int     `json:"stock"`
	Price float64 `json:"price"`
}

var inventory = map[int]InventoryItem{
	101: {ID: 101, Name: "Laptop", Stock: 50, Price: 999.99},
	102: {ID: 102, Name: "Mouse", Stock: 200, Price: 49.99},
	103: {ID: 103, Name: "Keyboard", Stock: 75, Price: 79.99},
}

func main() {
	ctx := context.Background()

	shutdown, err := telemetry.InitTracing(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("inventory-service")

	mux := http.NewServeMux()

	// /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "inventory-service",
		})
	})

	// Get all inventory
	mux.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "get_all_inventory")
		defer span.End()

		time.Sleep(time.Duration(rand.Intn(80)) * time.Millisecond)

		items := make([]InventoryItem, 0, len(inventory))
		for _, v := range inventory {
			items = append(items, v)
		}

		span.SetAttributes(
			attribute.Int("inventory.items_count", len(items)),
			attribute.String("operation.type", "read"),
		)

		json.NewEncoder(w).Encode(items)
		_ = ctx
	})

	// Get inventory by product
	mux.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "get_inventory_by_product")
		defer span.End()

		idStr := r.URL.Path[len("/inventory/"):]
		productID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid product id", http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.Int("product.id", productID),
			attribute.String("operation.type", "read"),
		)

		time.Sleep(time.Duration(rand.Intn(40)) * time.Millisecond)

		item, ok := inventory[productID]
		if !ok {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		span.SetAttributes(
			attribute.Int("product.stock", item.Stock),
			attribute.String("product.name", item.Name),
		)

		json.NewEncoder(w).Encode(item)
		_ = ctx
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: otelhttp.NewHandler(mux, "inventory-service"),
	}

	// Graceful shutdown
	go func() {
		log.Printf("Inventory service listening on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down inventory-service...")
	_ = server.Shutdown(ctx)
}
