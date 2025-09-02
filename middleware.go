package main

import (
	"fmt"
	"net/http"

	"github.com/pranayyb/whisp-rss-aggregator/internal/auth"
	"github.com/pranayyb/whisp-rss-aggregator/internal/db"
)

type authHandler func(http.ResponseWriter, *http.Request, db.User)

func (apiCfg *apiConfig) middlewareAuth(handler authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error getting API key: %v", err))
			return
		}

		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error getting user: %v", err))
			return
		}
		handler(w, r, user)
	}
}
