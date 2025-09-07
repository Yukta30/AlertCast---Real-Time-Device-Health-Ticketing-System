package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"alertcast/internal/cache"
	"alertcast/internal/config"
	"alertcast/internal/repository"
)

type Server struct {
	cfg *config.Config
	db  *repository.Postgres
	rc  *cache.Redis
}

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()

	db, err := repository.NewPostgres(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	rc := cache.New(cfg.RedisAddr)

	s := &Server{cfg: cfg, db: db, rc: rc}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/tickets/recent", s.handleRecentTickets)
	mux.HandleFunc("/stats", s.handleStats)

	addr := ":" + cfg.APIPort
	log.Printf("api: listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, withJSON(mux)))
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	type health struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	resp := health{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRecentTickets(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	tickets, err := s.db.RecentTickets(r.Context(), limit)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tickets)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	// Prefer DB counts to be canonical; Redis is best-effort cache.
	counts, err := s.db.SeverityCounts(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(counts)
}
