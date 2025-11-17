package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"subscriptions/internal/config"
	"subscriptions/internal/model"
	"subscriptions/internal/repo"
)

type SubscriptionHandler struct {
	repo *repo.SubscriptionRepo
}

func NewSubscriptionHandler(r *repo.SubscriptionRepo) *SubscriptionHandler {
	return &SubscriptionHandler{repo: r}
}

// RegisterSubscriptionRoutes connects to DB (using config) and registers routes on provided router.
// We keep a simple single-file helper: it will open db and create repo.
func RegisterSubscriptionRoutes(r chi.Router, cfg config.Config) {
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to DB in handler.RegisterSubscriptionRoutes")
	}
	h := NewSubscriptionHandler(repo.NewSubscriptionRepo(db))

	r.Post("/subscriptions", h.Create)
	r.Get("/subscriptions", h.List)
	r.Get("/subscriptions/{id}", h.Get)
	r.Put("/subscriptions/{id}", h.Update)
	r.Delete("/subscriptions/{id}", h.Delete)
	r.Get("/subscriptions/summary", h.Summary)
}

// Create - POST /subscriptions
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var s model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}
	// basic validation
	if strings.TrimSpace(s.ServiceName) == "" || s.Price <= 0 {
		http.Error(w, "service_name and positive price required", http.StatusBadRequest)
		return
	}
	if s.UserID == uuid.Nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	if s.StartMonth < 1 || s.StartMonth > 12 || s.StartYear <= 0 {
		http.Error(w, "invalid start month/year", http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(&s); err != nil {
		log.Error().Err(err).Msg("repo create failed")
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

// List - GET /subscriptions
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List()
	if err != nil {
		log.Error().Err(err).Msg("repo list failed")
		http.Error(w, "failed to list subscriptions", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []model.Subscription{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// Get - GET /subscriptions/{id}
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	s, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(s)
}

// Update - PUT /subscriptions/{id}
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var s model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	s.ID = id
	if err := h.repo.Update(&s); err != nil {
		log.Error().Err(err).Msg("update failed")
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(s)
}

// Delete - DELETE /subscriptions/{id}
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.repo.Delete(id); err != nil {
		log.Error().Err(err).Msg("delete failed")
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Summary - GET /subscriptions/summary?from=MM-YYYY&to=MM-YYYY&user_id=<uuid>&service_name=<text>
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "from and to required in MM-YYYY", http.StatusBadRequest)
		return
	}

	parseMMYYYY := func(s string) (int, int, error) {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return 0, 0, http.ErrNotSupported
		}
		m, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		y, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return m, y, nil
	}

	fromM, fromY, err := parseMMYYYY(from)
	if err != nil {
		http.Error(w, "invalid from format", http.StatusBadRequest)
		return
	}
	toM, toY, err := parseMMYYYY(to)
	if err != nil {
		http.Error(w, "invalid to format", http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &uid
	}

	var serviceName *string
	if s := r.URL.Query().Get("service_name"); s != "" {
		serviceName = &s
	}

	total, err := h.repo.Summary(fromM, fromY, toM, toY, userID, serviceName)
	if err != nil {
		log.Error().Err(err).Msg("summary failed")
		http.Error(w, "summary failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"total": total})
}
