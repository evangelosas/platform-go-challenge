package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPaginationAndBulk(t *testing.T) {
	store := NewInMemoryStore()
	s := NewServer(store)
	user := "u3"

	// bulk add two assets
	payload := []map[string]any{
		{
			"type":        "insight",
			"description": "a1",
			"payload":     map[string]any{"text": "hello"},
		},
		{
			"type":        "chart",
			"description": "a2",
			"payload":     map[string]any{"title": "t", "xAxisTitle": "x", "yAxisTitle": "y", "data": []float64{1}},
		},
	}
	w := doReq(s, http.MethodPost, "/users/"+user+"/favourites/bulk", payload)
	if w.Code != http.StatusCreated {
		t.Fatalf("bulk expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created []map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if len(created) != 2 {
		t.Fatalf("expected 2 created, got %d", len(created))
	}

	// pagination: limit 1 should return 1
	w = doReq(s, http.MethodGet, "/users/"+user+"/favourites?limit=1&offset=0", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
	var list []map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 item with limit=1, got %d", len(list))
	}
}

func TestPatchValidationAndDeleteNotFound(t *testing.T) {
	store := NewInMemoryStore()
	s := NewServer(store)
	user := "u4"

	// patch with empty description
	w := doReq(s, http.MethodPatch, "/users/"+user+"/favourites/aid1", map[string]string{"description": ""})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	// delete not found
	w = doReq(s, http.MethodDelete, "/users/"+user+"/favourites/nope", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
