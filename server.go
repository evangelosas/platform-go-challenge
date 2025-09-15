package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
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
