package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck godoc
// @Summary      Show the status of server.
// @Description  get the status of server.
// @Tags         health
// @Accept       */*
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
