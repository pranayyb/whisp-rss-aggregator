package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, status int, message string) {
	if status > 499 {
		log.Println("Server error:", message)
	}
	type errResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errResponse{
		Error: message,
	})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Panicln("Error marshalling JSON response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dat)
}
