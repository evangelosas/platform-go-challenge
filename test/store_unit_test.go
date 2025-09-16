package gwiExercise_test

import (
	"os"
	"path/filepath"
	"testing"

	gwiExercise "platform-go-challenge/gwiExercise"
)

func TestInMemoryStore_AddDuplicateAndErrors(t *testing.T) {
	s := gwiExercise.NewInMemoryStore()
	user := "u1"

	// add with empty ID assigns one
	chart := &gwiExercise.ChartAsset{BaseAsset: gwiExercise.BaseAsset{Type: gwiExercise.AssetChart, Description: "d1"}, Title: "t"}
	a, err := s.Add(user, chart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.GetID() == "" {
		t.Fatalf("expected id to be set")
	}

	// duplicate id for same user
	dup := &gwiExercise.ChartAsset{BaseAsset: gwiExercise.BaseAsset{ID: a.GetID(), Type: gwiExercise.AssetChart, Description: "d1"}, Title: "t"}
	if _, err := s.Add(user, dup); err == nil {
		t.Fatalf("expected duplicate error")
	}

	// remove non-existing user
	if err := s.Remove("nouser", "x"); err == nil {
		t.Fatalf("expected error on missing user")
	}
	// remove non-existing asset
	if err := s.Remove(user, "noasset"); err == nil {
		t.Fatalf("expected error on missing asset")
	}

	// update non-existing user/asset
	if _, err := s.UpdateDescription("nouser", "x", "d"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := s.UpdateDescription(user, "x", "d"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestFileStore_PersistReload(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fav.json")
	fs, err := gwiExercise.NewFileStore(file)
	if err != nil {
		t.Fatalf("init filestore: %v", err)
	}

	user := "u2"
	_, err = fs.Add(user, &gwiExercise.InsightAsset{BaseAsset: gwiExercise.BaseAsset{Type: gwiExercise.AssetInsight, Description: "ins"}, Text: "hello"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	// reload new instance
	fs2, err := gwiExercise.NewFileStore(file)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	list := fs2.List(user)
	if len(list) != 1 {
		t.Fatalf("expected 1 item after reload, got %d", len(list))
	}

	// cleanup
	_ = os.Remove(file)
}
