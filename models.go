package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type AssetType string

const (
	AssetChart    AssetType = "chart"
	AssetInsight  AssetType = "insight"
	AssetAudience AssetType = "audience"
)

type Asset interface {
	GetID() string
	GetType() AssetType
	GetDescription() string
	SetDescription(desc string)
	json.Marshaler
}

type BaseAsset struct {
	ID          string    `json:"id"`
	Type        AssetType `json:"type"`
	Description string    `json:"description"`
}

func (b *BaseAsset) GetID() string           { return b.ID }
func (b *BaseAsset) GetType() AssetType      { return b.Type }
func (b *BaseAsset) GetDescription() string  { return b.Description }
func (b *BaseAsset) SetDescription(d string) { b.Description = d }

// ChartAsset represents a chart with basic metadata and numeric data points.
type ChartAsset struct {
	BaseAsset
	Title      string    `json:"title"`
	XAxisTitle string    `json:"xAxisTitle"`
	YAxisTitle string    `json:"yAxisTitle"`
	Data       []float64 `json:"data"`
}

func (c *ChartAsset) MarshalJSON() ([]byte, error) {
	type alias ChartAsset
	return json.Marshal((*alias)(c))
}

// InsightAsset represents a short insight text.
type InsightAsset struct {
	BaseAsset
	Text string `json:"text"`
}

func (i *InsightAsset) MarshalJSON() ([]byte, error) {
	type alias InsightAsset
	return json.Marshal((*alias)(i))
}

// AudienceAsset represents a simple audience definition for the exercise.
type AudienceAsset struct {
	BaseAsset
	Gender             string `json:"gender"`
	BirthCountry       string `json:"birthCountry"`
	AgeGroup           string `json:"ageGroup"`
	SocialHoursPerDay  string `json:"socialHoursPerDay"`
	PurchasesLastMonth int    `json:"purchasesLastMonth"`
}

func (a *AudienceAsset) MarshalJSON() ([]byte, error) {
	type alias AudienceAsset
	return json.Marshal((*alias)(a))
}

// DecodeAssetFactory parses an incoming payload into the proper asset type.
// On add we allow omitting ID; callers can provide it optionally.
type AddAssetRequest struct {
	ID          string          `json:"id"`
	Type        AssetType       `json:"type"`
	Description string          `json:"description"`
	Payload     json.RawMessage `json:"payload"`
}

func (r AddAssetRequest) ValidateBasic() error {
	if r.Type == "" {
		return errors.New("type is required")
	}
	if r.Description == "" {
		return errors.New("description is required")
	}
	if len(r.Payload) == 0 {
		return errors.New("payload is required")
	}
	return nil
}

func DecodeAssetFromAddRequest(req AddAssetRequest) (Asset, error) {
	switch req.Type {
	case AssetChart:
		var body struct {
			Title      string    `json:"title"`
			XAxisTitle string    `json:"xAxisTitle"`
			YAxisTitle string    `json:"yAxisTitle"`
			Data       []float64 `json:"data"`
		}
		if err := json.Unmarshal(req.Payload, &body); err != nil {
			return nil, fmt.Errorf("invalid chart payload: %w", err)
		}
		asset := &ChartAsset{BaseAsset: BaseAsset{ID: req.ID, Type: AssetChart, Description: req.Description},
			Title: body.Title, XAxisTitle: body.XAxisTitle, YAxisTitle: body.YAxisTitle, Data: body.Data}
		return asset, nil
	case AssetInsight:
		var body struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(req.Payload, &body); err != nil {
			return nil, fmt.Errorf("invalid insight payload: %w", err)
		}
		asset := &InsightAsset{BaseAsset: BaseAsset{ID: req.ID, Type: AssetInsight, Description: req.Description}, Text: body.Text}
		return asset, nil
	case AssetAudience:
		var body struct {
			Gender             string `json:"gender"`
			BirthCountry       string `json:"birthCountry"`
			AgeGroup           string `json:"ageGroup"`
			SocialHoursPerDay  string `json:"socialHoursPerDay"`
			PurchasesLastMonth int    `json:"purchasesLastMonth"`
		}
		if err := json.Unmarshal(req.Payload, &body); err != nil {
			return nil, fmt.Errorf("invalid audience payload: %w", err)
		}
		asset := &AudienceAsset{BaseAsset: BaseAsset{ID: req.ID, Type: AssetAudience, Description: req.Description},
			Gender: body.Gender, BirthCountry: body.BirthCountry, AgeGroup: body.AgeGroup, SocialHoursPerDay: body.SocialHoursPerDay, PurchasesLastMonth: body.PurchasesLastMonth}
		return asset, nil
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", req.Type)
	}
}
