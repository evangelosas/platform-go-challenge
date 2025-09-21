package gwiexercise

import (
	"errors"
	"testing"
	"time"
)

type mockAsset struct {
	id, description string
	createdAt       time.Time
}

func (m *mockAsset) GetID() string                { return m.id }
func (m *mockAsset) GetType() AssetType           { return "" }
func (m *mockAsset) GetDescription() string       { return m.description }
func (m *mockAsset) SetDescription(desc string)   { m.description = desc }
func (m *mockAsset) MarshalJSON() ([]byte, error) { return nil, nil }

func TestInMemoryStore_List(t *testing.T) {
	var store Store = NewInMemoryStore()
	store.Add("user1", &mockAsset{id: "asset1", description: "desc1"})
	store.Add("user1", &mockAsset{id: "asset2", description: "desc2"})

	tests := []struct {
		name   string
		userID string
		want   []Asset
	}{
		{
			name:   "existing user",
			userID: "user1",
			want:   []Asset{&mockAsset{id: "asset1", description: "desc1"}, &mockAsset{id: "asset2", description: "desc2"}},
		},
		{
			name:   "non-existing user",
			userID: "user2",
			want:   []Asset{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.List(tt.userID)
			if len(got) != len(tt.want) {
				t.Errorf("expected %d assets, got %d", len(tt.want), len(got))
			}
			for i, asset := range got {
				if asset.GetID() != tt.want[i].GetID() || asset.GetDescription() != tt.want[i].GetDescription() {
					t.Errorf("expected asset %v, got %v", tt.want[i], asset)
				}
			}
		})
	}
}

func TestInMemoryStore_Add(t *testing.T) {
	var store Store = NewInMemoryStore()

	tests := []struct {
		name     string
		userID   string
		asset    Asset
		expected error
	}{
		{
			name:     "add new asset",
			userID:   "user1",
			asset:    &mockAsset{id: "asset1", description: "desc1"},
			expected: nil,
		},
		{
			name:     "duplicate asset ID",
			userID:   "user1",
			asset:    &mockAsset{id: "asset1", description: "desc2"},
			expected: errors.New("asset with same id already exists for user"),
		},
		{
			name:     "new asset for different user",
			userID:   "user2",
			asset:    &mockAsset{id: "asset2", description: "desc3"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.Add(tt.userID, tt.asset)
			if (err != nil) != (tt.expected != nil) {
				t.Errorf("received error %v, expected %v", err, tt.expected)
			}
		})
	}
}

func TestInMemoryStore_Remove(t *testing.T) {
	var store Store = NewInMemoryStore()
	store.Add("user1", &mockAsset{id: "asset1", description: "desc1"})

	tests := []struct {
		name     string
		userID   string
		assetID  string
		expected error
	}{
		{
			name:     "remove existing asset",
			userID:   "user1",
			assetID:  "asset1",
			expected: nil,
		},
		{
			name:     "remove non-existing asset",
			userID:   "user1",
			assetID:  "asset2",
			expected: errors.New("asset not found"),
		},
		{
			name:     "remove asset for non-existing user",
			userID:   "user2",
			assetID:  "asset1",
			expected: errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Remove(tt.userID, tt.assetID)
			if (err != nil) != (tt.expected != nil) || (err != nil && err.Error() != tt.expected.Error()) {
				t.Errorf("unexpected error. got %v, want %v", err, tt.expected)
			}
		})
	}
}

func TestInMemoryStore_UpdateDescription(t *testing.T) {
	var store Store = NewInMemoryStore()
	store.Add("user1", &mockAsset{id: "asset1", description: "desc1"})

	tests := []struct {
		name        string
		userID      string
		assetID     string
		description string
		expected    error
	}{
		{
			name:        "update existing asset description",
			userID:      "user1",
			assetID:     "asset1",
			description: "new desc",
			expected:    nil,
		},
		{
			name:        "update non-existing asset",
			userID:      "user1",
			assetID:     "asset2",
			description: "desc",
			expected:    errors.New("asset not found"),
		},
		{
			name:        "update asset for non-existing user",
			userID:      "user2",
			assetID:     "asset1",
			description: "desc",
			expected:    errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.UpdateDescription(tt.userID, tt.assetID, tt.description)
			if (err != nil) != (tt.expected != nil) || (err != nil && err.Error() != tt.expected.Error()) {
				t.Errorf("unexpected error. got %v, want %v", err, tt.expected)
			}
		})
	}
}
