package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type ProviderCache struct {
	root string
}

type cachedAnalysisResult struct {
	Scenes     []models.Scene  `json:"scenes"`
	Metadata   models.Metadata `json:"metadata,omitempty"`
	PayloadRef string          `json:"payload_ref,omitempty"`
}

type cachedSlotTemplate struct {
	Rank                  int             `json:"rank"`
	SceneNumber           int             `json:"scene_number"`
	AnchorStartFrame      int             `json:"anchor_start_frame"`
	AnchorEndFrame        int             `json:"anchor_end_frame"`
	QuietWindowSeconds    float64         `json:"quiet_window_seconds"`
	Score                 float64         `json:"score"`
	Reasoning             string          `json:"reasoning"`
	ContextRelevanceScore *float64        `json:"context_relevance_score,omitempty"`
	NarrativeFitScore     *float64        `json:"narrative_fit_score,omitempty"`
	AnchorContinuityScore *float64        `json:"anchor_continuity_score,omitempty"`
	Metadata              models.Metadata `json:"metadata,omitempty"`
}

type cachedGenerationOutput struct {
	GeneratedClipPath  string          `json:"generated_clip_path"`
	GeneratedAudioPath string          `json:"generated_audio_path"`
	Metadata           models.Metadata `json:"metadata,omitempty"`
}

func NewProviderCache(root string) *ProviderCache {
	return &ProviderCache{root: strings.TrimSpace(root)}
}

func (c *ProviderCache) Enabled() bool {
	return c != nil && c.root != ""
}

func (c *ProviderCache) HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read cache file input %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func (c *ProviderCache) ProductFingerprint(product models.Product) (string, error) {
	payload := map[string]any{
		"name":             strings.TrimSpace(product.Name),
		"description":      strings.TrimSpace(product.Description),
		"category":         strings.TrimSpace(product.Category),
		"context_keywords": append([]string(nil), product.ContextKeywords...),
		"source_url":       strings.TrimSpace(product.SourceURL),
	}
	if product.ImagePath != "" {
		imageHash, err := c.HashFile(product.ImagePath)
		if err != nil {
			return "", err
		}
		payload["image_hash"] = imageHash
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal product fingerprint: %w", err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (c *ProviderCache) LoadAnalysis(videoHash string) (cachedAnalysisResult, bool, error) {
	var result cachedAnalysisResult
	ok, err := c.loadJSON(c.analysisPath(videoHash), &result)
	return result, ok, err
}

func (c *ProviderCache) SaveAnalysis(videoHash string, result cachedAnalysisResult) error {
	return c.saveJSON(c.analysisPath(videoHash), result)
}

func (c *ProviderCache) LoadRanking(videoHash, productHash string) ([]cachedSlotTemplate, bool, error) {
	var result []cachedSlotTemplate
	ok, err := c.loadJSON(c.rankingPath(videoHash, productHash), &result)
	return result, ok, err
}

func (c *ProviderCache) SaveRanking(videoHash, productHash string, templates []cachedSlotTemplate) error {
	return c.saveJSON(c.rankingPath(videoHash, productHash), templates)
}

func (c *ProviderCache) LoadSuggestedLine(key string) (string, bool, error) {
	var payload struct {
		Value string `json:"value"`
	}
	ok, err := c.loadJSON(c.promptPath("suggested-line", key), &payload)
	return payload.Value, ok, err
}

func (c *ProviderCache) SaveSuggestedLine(key, value string) error {
	return c.saveJSON(c.promptPath("suggested-line", key), map[string]string{"value": value})
}

func (c *ProviderCache) LoadGenerationBrief(key string) (string, bool, error) {
	var payload struct {
		Value string `json:"value"`
	}
	ok, err := c.loadJSON(c.promptPath("generation-brief", key), &payload)
	return payload.Value, ok, err
}

func (c *ProviderCache) SaveGenerationBrief(key, value string) error {
	return c.saveJSON(c.promptPath("generation-brief", key), map[string]string{"value": value})
}

func (c *ProviderCache) LoadGenerationOutput(key string) (cachedGenerationOutput, bool, error) {
	var result cachedGenerationOutput
	ok, err := c.loadJSON(c.promptPath("generation-output", key), &result)
	return result, ok, err
}

func (c *ProviderCache) SaveGenerationOutput(key string, result cachedGenerationOutput) error {
	return c.saveJSON(c.promptPath("generation-output", key), result)
}

func (c *ProviderCache) PromptKey(parts ...string) string {
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized = append(normalized, strings.TrimSpace(part))
	}
	body, _ := json.Marshal(normalized)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func buildCachedSlotTemplates(slots []models.Slot, scenes []models.Scene) []cachedSlotTemplate {
	sceneNumbers := make(map[string]int, len(scenes))
	for _, scene := range scenes {
		sceneNumbers[scene.ID] = scene.SceneNumber
	}
	templates := make([]cachedSlotTemplate, 0, len(slots))
	for _, slot := range slots {
		template := cachedSlotTemplate{
			Rank:                  slot.Rank,
			SceneNumber:           sceneNumbers[slot.SceneID],
			AnchorStartFrame:      slot.AnchorStartFrame,
			AnchorEndFrame:        slot.AnchorEndFrame,
			QuietWindowSeconds:    slot.QuietWindowSeconds,
			Score:                 slot.Score,
			Reasoning:             slot.Reasoning,
			ContextRelevanceScore: slot.ContextRelevanceScore,
			NarrativeFitScore:     slot.NarrativeFitScore,
			AnchorContinuityScore: slot.AnchorContinuityScore,
			Metadata:              cloneMetadata(slot.Metadata),
		}
		templates = append(templates, template)
	}
	sort.SliceStable(templates, func(i, j int) bool {
		return templates[i].Rank < templates[j].Rank
	})
	return templates
}

func rehydrateCachedSlots(jobID string, sourceFPS float64, scenes []models.Scene, templates []cachedSlotTemplate) []models.Slot {
	sceneByNumber := make(map[int]models.Scene, len(scenes))
	for _, scene := range scenes {
		sceneByNumber[scene.SceneNumber] = scene
	}
	now := TimestampNow()
	slots := make([]models.Slot, 0, len(templates))
	for _, template := range templates {
		scene, ok := sceneByNumber[template.SceneNumber]
		if !ok {
			continue
		}
		slotID := fmt.Sprintf("slot_%s_%s_%d_%d", jobID, scene.ID, template.AnchorStartFrame, template.AnchorEndFrame)
		slots = append(slots, models.Slot{
			ID:                    slotID,
			JobID:                 jobID,
			Rank:                  template.Rank,
			SceneID:               scene.ID,
			AnchorStartFrame:      template.AnchorStartFrame,
			AnchorEndFrame:        template.AnchorEndFrame,
			SourceFPS:             sourceFPS,
			QuietWindowSeconds:    template.QuietWindowSeconds,
			Score:                 template.Score,
			Reasoning:             template.Reasoning,
			Status:                constants.SlotStatusProposed,
			ContextRelevanceScore: template.ContextRelevanceScore,
			NarrativeFitScore:     template.NarrativeFitScore,
			AnchorContinuityScore: template.AnchorContinuityScore,
			Metadata:              cloneMetadata(template.Metadata),
			CreatedAt:             now,
			UpdatedAt:             now,
		})
	}
	return slots
}

func cacheSceneKey(scene models.Scene) string {
	return fmt.Sprintf("%d:%.3f:%.3f", scene.SceneNumber, scene.StartSeconds, scene.EndSeconds)
}

func cacheSlotKey(slot models.Slot) string {
	return fmt.Sprintf("%d:%d", slot.AnchorStartFrame, slot.AnchorEndFrame)
}

func (c *ProviderCache) analysisPath(videoHash string) string {
	return filepath.Join(c.root, "analysis", videoHash+".json")
}

func (c *ProviderCache) rankingPath(videoHash, productHash string) string {
	return filepath.Join(c.root, "ranking", videoHash, productHash+".json")
}

func (c *ProviderCache) promptPath(kind, key string) string {
	return filepath.Join(c.root, kind, key+".json")
}

func (c *ProviderCache) loadJSON(path string, target any) (bool, error) {
	if !c.Enabled() {
		return false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read cache file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("decode cache file %s: %w", path, err)
	}
	return true, nil
}

func (c *ProviderCache) saveJSON(path string, value any) error {
	if !c.Enabled() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache directory for %s: %w", path, err)
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache file %s: %w", path, err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write cache file %s: %w", path, err)
	}
	return nil
}
