package gwiTest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	store Store
}

func NewServer(store Store) *Server {
	return &Server{store: store}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.Trim(r.URL.Path, "/")
	segs := strings.Split(path, "/")
	if len(segs) < 3 || segs[0] != "users" || segs[2] != "favourites" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	userID := segs[1]

	if len(segs) == 3 {
		// /users/{userID}/favourites
		s.handleUserFavourites(w, r, userID)
		return
	}
	if len(segs) == 4 {
		if segs[3] == "bulk" {
			s.handleUserFavouritesBulk(w, r, userID)
			return
		}
		assetID := segs[3]
		s.handleUserFavouriteByID(w, r, userID, assetID)
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

func (s *Server) handleUserFavourites(w http.ResponseWriter, r *http.Request, userID string) {
	switch r.Method {
	case http.MethodGet:
		assets := s.store.List(userID)
		// sorting
		q := r.URL.Query()
		sortField := q.Get("sort")
		if sortField == "" {
			sortField = "created_at"
		}
		order := strings.ToLower(q.Get("order"))
		desc := order == "desc"
		sort.Slice(assets, func(i, j int) bool {
			a, b := assets[i], assets[j]
			switch sortField {
			case "type":
				if desc {
					return string(a.GetType()) > string(b.GetType())
				}
				return string(a.GetType()) < string(b.GetType())
			case "created_at":
				ai, aj := createdAtOf(a), createdAtOf(b)
				if desc {
					return ai.After(aj)
				}
				return ai.Before(aj)
			default:
				ai, aj := createdAtOf(a), createdAtOf(b)
				if desc {
					return ai.After(aj)
				}
				return ai.Before(aj)
			}
		})
		// pagination
		limit := parseIntDefault(q.Get("limit"), 100)
		if limit <= 0 || limit > 1000 {
			limit = 100
		}
		offset := parseIntDefault(q.Get("offset"), 0)
		if offset < 0 {
			offset = 0
		}
		end := offset + limit
		if offset > len(assets) {
			assets = []Asset{}
		} else if end < len(assets) {
			assets = assets[offset:end]
		} else if offset < len(assets) {
			assets = assets[offset:]
		}
		writeJSON(w, http.StatusOK, assets)
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		var req AddAssetRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if err := req.ValidateBasic(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		asset, err := DecodeAssetFromAddRequest(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := s.store.Add(userID, asset)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleUserFavouritesBulk(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var reqs []AddAssetRequest
	if err := json.Unmarshal(body, &reqs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json array: "+err.Error())
		return
	}
	assets := make([]Asset, 0, len(reqs))
	for i, r := range reqs {
		if err := r.ValidateBasic(); err != nil {
			writeError(w, http.StatusBadRequest, "item "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		a, err := DecodeAssetFromAddRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "item "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		assets = append(assets, a)
	}
	created := make([]Asset, 0, len(assets))
	for i, a := range assets {
		c, err := s.store.Add(userID, a)
		if err != nil {
			writeError(w, http.StatusConflict, "item "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		created = append(created, c)
	}
	writeJSON(w, http.StatusCreated, created)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func createdAtOf(a Asset) time.Time {
	switch v := a.(type) {
	case *ChartAsset:
		return v.CreatedAt
	case *InsightAsset:
		return v.CreatedAt
	case *AudienceAsset:
		return v.CreatedAt
	default:
		return time.Time{}
	}
}

func (s *Server) handleUserFavouriteByID(w http.ResponseWriter, r *http.Request, userID, assetID string) {
	switch r.Method {
	case http.MethodDelete:
		if err := s.store.Remove(userID, assetID); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPatch:
		var req struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if strings.TrimSpace(req.Description) == "" {
			writeError(w, http.StatusBadRequest, "description is required")
			return
		}
		updated, err := s.store.UpdateDescription(userID, assetID, req.Description)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)
	default:
		w.Header().Set("Allow", "PATCH, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		log.Printf("write json error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
