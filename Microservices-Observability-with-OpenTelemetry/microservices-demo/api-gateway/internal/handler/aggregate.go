package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"api-gateway/internal/client"
)

func AggregateUserOrders(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("api-gateway")

	ctx, span := tracer.Start(r.Context(), "get_user_orders_aggregate")
	defer span.End()

	// /api/users/{id}/orders
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}

	userID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int("user.id", userID),
		attribute.String("operation.type", "aggregate"),
	)

	user := client.GetUser(ctx, userID)
	result := map[string]interface{}{
		"user": user,
	}

	json.NewEncoder(w).Encode(result)
}
