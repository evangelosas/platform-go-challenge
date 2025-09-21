package gwiexercise_test

import (
	"net/http"
	"path/filepath"
	"testing"

	gwiExercise "platform-go-challenge/gwiexercise"
)

// Benchmark for the in-memory store.
func BenchmarkInMemoryAddList(b *testing.B) {
	store := gwiExercise.NewInMemoryStore()
	benchAddList(b, store, "us")
}

// BenchmarkInFileAddList for the file store.
func BenchmarkInFileAddList(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "favourites_bench.json")
	store, err := gwiExercise.NewFileStore(path)
	if err != nil {
		b.Fatalf("NewFileStore error: %v", err)
	}
	benchAddList(b, store, "uf")
}

func benchAddList(b *testing.B, store gwiExercise.Store, user string) {
	for i := 0; i < b.N; i++ {
		_, _ = store.Add(user, &gwiExercise.InsightAsset{BaseAsset: gwiExercise.BaseAsset{Type: gwiExercise.AssetInsight, Description: "d"}, Text: "x"})
	}
	b.ReportAllocs()
	_ = store.List(user)
}

func BenchmarkInMemoryBulkEndpoint(b *testing.B) {
	store := gwiExercise.NewInMemoryStore()
	s := gwiExercise.NewServer(store)
	user := "ub2"
	payload := make([]map[string]any, 0, 100)
	for i := 0; i < 100; i++ {
		payload = append(payload, map[string]any{"type": "insight", "description": "d", "payload": map[string]any{"text": "hello"}})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := doReq(s, http.MethodPost, "/users/"+user+"/favourites/bulk", payload)
		if w.Code != http.StatusCreated {
			b.Fatalf("expected 201, got %d", w.Code)
		}
	}
}
