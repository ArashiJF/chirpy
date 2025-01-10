package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) create_user(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashed})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating user", err)
		return
	}

	responseWithJSON(w, http.StatusCreated, successS{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) login_user(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type successS struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	passErr := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if passErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	// fmt.Printf("Token created: %s\n", token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh JWT", err)
		return
	}

	_, createRefreshErr := cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
		ExpiresAt: sql.NullTime{
			Time:  time.Now().UTC().Add(time.Hour * 24 * 60), // 60 days
			Valid: true,
		},
	})
	if createRefreshErr != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	responseWithJSON(w, 200, successS{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) refresh_token(w http.ResponseWriter, req *http.Request) {
	headerToken, bearerErr := auth.GetBearerToken(req.Header)
	if bearerErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Bad bearer", bearerErr)
		return
	}

	type successS struct {
		Token string `json:"token"`
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), headerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Expired token", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	responseWithJSON(w, 200, successS{
		Token: token,
	})
}

func (cfg *apiConfig) revoke_refresh(w http.ResponseWriter, req *http.Request) {
	headerToken, bearerErr := auth.GetBearerToken(req.Header)
	if bearerErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Bad bearer", bearerErr)
		return
	}

	_, err := cfg.db.RevokeRefreshToken(req.Context(), database.RevokeRefreshTokenParams{
		Token: headerToken,
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Expired token", err)
		return
	}

	responseWithJSON(w, 204, nil)
}
