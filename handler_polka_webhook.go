package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/vmamchur/go_chirpy-api/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get API key", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Wrong API key", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse userID", err)
		return
	}
	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
