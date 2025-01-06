package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func censor_words(text string, forbidden_words []string) string {
	replacement := "****"
	aux := text
	curr := 0
	max := len(forbidden_words)

	for {
		if curr < max {
			target := strings.Index(strings.ToLower(aux), forbidden_words[curr])
			if target != -1 {
				aux = aux[:target] + replacement + aux[target+len(forbidden_words[curr]):]
			} else {
				curr += 1
			}
		} else {
			break
		}
	}
	return aux
}

func validate_length(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type successS struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	forbidden_words := []string{"kerfuffle", "sharbert", "fornax"}

	responseWithJSON(w, http.StatusOK, successS{
		CleanedBody: censor_words(params.Body, forbidden_words),
	})
}

func (cfg *apiConfig) create_user(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type successS struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "oops", err)
		return
	}

	responseWithJSON(w, http.StatusCreated, successS{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
