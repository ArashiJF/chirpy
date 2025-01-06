package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}
	forbidden_words := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned := censor_words(body, forbidden_words)
	return cleaned, nil
}

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

func (cfg *apiConfig) create_chirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type successS struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{Body: cleanedBody, UserID: params.UserID})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not create chirp", err)
		return
	}

	responseWithJSON(w, http.StatusCreated, successS{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
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
