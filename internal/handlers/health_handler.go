package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	db *mongo.Client
}

func NewHealthHandler(db *mongo.Client) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	// Check database connection
	if h.db != nil {
		ctx := context.Background()
		err := h.db.Ping(ctx, nil)
		if err != nil {
			response.Status = "error"
			response.Database = "disconnected"
			response.Error = err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			response.Database = "connected"
		}
	} else {
		response.Status = "error"
		response.Database = "not configured"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

