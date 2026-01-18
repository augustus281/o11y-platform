package handler

import (
	"encoding/json"
	"net/http"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "api-gateway",
	})
}
