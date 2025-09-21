package gwiexercise

import (
	"sort"
	"testing"
	"time"
)

func TestInMemoryStore_List_EmptyForNewUser(t *testing.T) {
	s := NewInMemoryStore()
	if got := s.List("nouser"); len(got) != 0 {
		t.Fatalf("expected empty list, got %d", len(got))
	}
}

func TestInMemoryStore_Add_AssignsID_AndSetsCreatedAtIfZero(t *testing.T) {
	s := NewInMemoryStore()

	// CreatedAt should be set if zero and ID should be auto-assigned if empty
	chart := &ChartAsset{BaseAsset: BaseAsset{Type: AssetChart, Description: "desc"}, Title: "T"}
	added, err := s.Add("u1", chart)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if added.GetID() == "" {
		t.Fatalf("expected ID to be assigned")
	}
	if chart.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}

	// If ID is supplied, it should be respected
	ins := &InsightAsset{BaseAsset: BaseAsset{ID: "given-id", Type: AssetInsight, Description: "d"}, Text: "hi"}
	added2, err := s.Add("u1", ins)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if added2.GetID() != "given-id" {
		t.Fatalf("expected given id to be preserved, got %q", added2.GetID())
	}
}

func TestInMemoryStore_Add_DuplicateAndPerUserIsolation(t *testing.T) {
	s := NewInMemoryStore()

	a := &AudienceAsset{BaseAsset: BaseAsset{ID: "same", Type: AssetAudience, Description: "da"}}
	if _, err := s.Add("u1", a); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}
	// duplicate in same user
	if _, err := s.Add("u1", &AudienceAsset{BaseAsset: BaseAsset{ID: "same", Type: AssetAudience, Description: "db"}}); err == nil || err.Error() != "asset with same id already exists for user" {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	// same ID allowed for different user
	if _, err := s.Add("u2", &AudienceAsset{BaseAsset: BaseAsset{ID: "same", Type: AssetAudience, Description: "dc"}}); err != nil {
		t.Fatalf("unexpected error for different user: %v", err)
	}
}

func TestInMemoryStore_List_OrderingIrrelevant(t *testing.T) {
	s := NewInMemoryStore()

	s.Add("u", &InsightAsset{BaseAsset: BaseAsset{ID: "b", Type: AssetInsight, Description: "d"}, Text: "t"})
	s.Add("u", &InsightAsset{BaseAsset: BaseAsset{ID: "a", Type: AssetInsight, Description: "d"}, Text: "t"})

	got := s.List("u")
	if len(got) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(got))
	}
	ids := []string{got[0].GetID(), got[1].GetID()}
	sort.Strings(ids)
	if !(ids[0] == "a" && ids[1] == "b") {
		t.Fatalf("unexpected ids: %v", ids)
	}
}

func TestInMemoryStore_Remove(t *testing.T) {
	s := NewInMemoryStore()

	s.Add("u", &ChartAsset{BaseAsset: BaseAsset{ID: "x", Type: AssetChart, Description: "d"}, Title: "t"})

	if err := s.Remove("u", "x"); err != nil {
		t.Fatalf("remove existing: %v", err)
	}
	if err := s.Remove("u", "missing"); err == nil || err.Error() != "asset not found" {
		t.Fatalf("expected asset not found, got %v", err)
	}
	if err := s.Remove("nouser", "x"); err == nil || err.Error() != "user not found" {
		t.Fatalf("expected user not found, got %v", err)
	}
}

func TestInMemoryStore_UpdateDescription(t *testing.T) {
	s := NewInMemoryStore()

	s.Add("u", &InsightAsset{BaseAsset: BaseAsset{ID: "id1", Type: AssetInsight, Description: "old"}, Text: "t"})

	updated, err := s.UpdateDescription("u", "id1", "new")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.GetDescription() != "new" {
		t.Fatalf("description not updated: %v", updated.GetDescription())
	}

	if _, err := s.UpdateDescription("nouser", "id1", "x"); err == nil || err.Error() != "user not found" {
		t.Fatalf("expected user not found, got %v", err)
	}
	if _, err := s.UpdateDescription("u", "missing", "x"); err == nil || err.Error() != "asset not found" {
		t.Fatalf("expected asset not found, got %v", err)
	}
}

func TestHelpers_base36_fmtID(t *testing.T) {
	// base36 known values
	cases := []struct {
		in  uint64
		out string
	}{
		{0, "0"},
		{1, "1"},
		{10, "a"},
		{35, "z"},
		{36, "10"},
		{71, "1z"},
		{1295, "zz"},
	}
	for _, c := range cases {
		if got := base36(c.in); got != c.out {
			t.Fatalf("base36(%d) = %q, want %q", c.in, got, c.out)
		}
	}

	// fmtID should concatenate two base36 values
	x := int64(36) // base36 -> "10"
	y := int64(71) // base36 -> "1z"
	if got := fmtID(x, y); got != "101z" {
		t.Fatalf("fmtID unexpected: %q", got)
	}
}

func Test_setCreatedAtIfZero_onlySetsWhenZero(t *testing.T) {
	s := NewInMemoryStore()
	// Add one with existing CreatedAt, ensure it remains
	ts := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	chart := &ChartAsset{BaseAsset: BaseAsset{ID: "keep", Type: AssetChart, Description: "d", CreatedAt: ts}, Title: "t"}
	if _, err := s.Add("u", chart); err != nil {
		t.Fatalf("add: %v", err)
	}
	if chart.CreatedAt != ts {
		t.Fatalf("CreatedAt should remain unchanged when already set")
	}
}
