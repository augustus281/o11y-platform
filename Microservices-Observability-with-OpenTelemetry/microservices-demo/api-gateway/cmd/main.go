package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"api-gateway/internal/handler"
	"api-gateway/internal/proxy"
	"api-gateway/internal/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	shutdown, err := telemetry.InitTracing(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/user/", handler.AggregateUserOrders)

	mux.Handle("/api/users/", proxy.New("/api/users", "http://user-service:3001"))
	mux.Handle("/api/orders/", proxy.New("/api/orders", "http://order-service:3002"))
	mux.Handle("/api/inventory/", proxy.New("/api/inventory", "http://inventory-service:3003"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: otelhttp.NewHandler(mux, "api-gateway"),
	}

	go func() {
		log.Printf("API Gateway listening on %s\n", port)
		_ = server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down API Gateway...")
	_ = server.Shutdown(ctx)
}
