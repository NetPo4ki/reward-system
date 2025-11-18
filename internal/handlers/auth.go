package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NetPo4ki/reward-system/internal/auth"
	"github.com/NetPo4ki/reward-system/internal/repo"
)

type AuthHandler struct {
	Users     *repo.UsersRepo
	JWTSecret string
	TokenTTL  time.Duration
}

type signupRequest struct {
	Username string `json:"username"`
}

type signupResponse struct {
	Token string `json:"token"`
	User  struct {
		ID         int64  `json:"id"`
		Username   string `json:"username"`
		ReferrerID *int64 `json:"referrer_id,omitempty"`
	} `json:"user"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	u, err := h.Users.Create(r.Context(), req.Username)
	if err != nil {
		http.Error(w, `{"error":"cannot create user"}`, http.StatusBadRequest)
		return
	}

	tok, err := auth.SignToken(u.ID, h.JWTSecret, h.TokenTTL)
	if err != nil {
		http.Error(w, `{"error":"cannot sign token"}`, http.StatusInternalServerError)
		return
	}

	var resp signupResponse
	resp.Token = tok
	resp.User.ID = u.ID
	resp.User.Username = u.Username
	resp.User.ReferrerID = u.ReferrerID

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
