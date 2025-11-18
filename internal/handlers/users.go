package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/NetPo4ki/reward-system/internal/auth"
	"github.com/NetPo4ki/reward-system/internal/models"
	"github.com/NetPo4ki/reward-system/internal/repo"
)

type UsersHandler struct {
	Users     *repo.UsersRepo
	Tasks     *repo.TasksRepo
	UserTasks *repo.UserTasksRepo
}

type statusResponse struct {
	User struct {
		ID         int64  `json:"id"`
		Username   string `json:"username"`
		ReferrerID *int64 `json:"referrer_id,omitempty"`
		CreatedAt  string `json:"created_at"`
	} `json:"user"`
	Balance   int                `json:"balance"`
	Completed []completedSummary `json:"completed"`
}

type completedSummary struct {
	TaskID      int64  `json:"task_id"`
	Points      int    `json:"points"`
	CompletedAt string `json:"completed_at"`
}

func (h *UsersHandler) Status(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pathID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id", "bad_request")
		return
	}
	if uid != pathID {
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	u, err := h.Users.GetByID(r.Context(), pathID)
	if err != nil {
		if err == repo.ErrNotFound {
			writeError(w, http.StatusNotFound, "user not found", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	balance, err := h.Users.Balance(r.Context(), pathID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	list, err := h.UserTasks.ListCompleted(r.Context(), pathID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	var resp statusResponse
	resp.User.ID = u.ID
	resp.User.Username = u.Username
	resp.User.ReferrerID = u.ReferrerID
	resp.User.CreatedAt = u.CreatedAt.UTC().Format(time.RFC3339)
	resp.Balance = balance
	for _, it := range list {
		resp.Completed = append(resp.Completed, completedSummary{
			TaskID:      it.TaskID,
			Points:      it.PointsAwarded,
			CompletedAt: it.CompletedAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

type leaderboardResponse struct {
	Items []models.LeaderboardEntry `json:"items"`
}

func (h *UsersHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 50, 1, 100)
	items, err := h.Users.Leaderboard(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, leaderboardResponse{Items: items})
}

type completeTaskRequest struct {
	TaskCode string `json:"task_code"`
}

type completeTaskResponse struct {
	NewlyCompleted bool `json:"newly_completed"`
}

func (h *UsersHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pathID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id", "bad_request")
		return
	}
	if authUserID != pathID {
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	var req completeTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TaskCode == "" {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	okNew, err := h.UserTasks.CompleteTask(r.Context(), pathID, req.TaskCode)
	if err != nil {
		if err == repo.ErrNotFound {
			writeError(w, http.StatusNotFound, "task not found or inactive", "not_found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, completeTaskResponse{NewlyCompleted: okNew})
}

type setReferrerRequest struct {
	ReferrerID int64 `json:"referrer_id"`
}

func (h *UsersHandler) SetReferrer(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	pathID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id", "bad_request")
		return
	}
	if authUserID != pathID {
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	var req setReferrerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReferrerID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if req.ReferrerID == pathID {
		writeError(w, http.StatusBadRequest, "cannot refer self", "invalid")
		return
	}

	if err := h.Users.SetReferrer(r.Context(), pathID, req.ReferrerID); err != nil {
		switch err {
		case repo.ErrNotFound:
			writeError(w, http.StatusNotFound, "referrer not found", "not_found")
		case repo.ErrConflict:
			writeError(w, http.StatusConflict, "referrer already set", "conflict")
		case repo.ErrInvalid:
			writeError(w, http.StatusBadRequest, "invalid referrer", "invalid")
		default:
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseLimit(r *http.Request, def, min, max int) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return def
	}
	v, err := strconv.Atoi(q)
	if err != nil {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
