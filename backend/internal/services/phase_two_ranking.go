package services

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const internalSlotRankingRequestIDKey = "slot_ranking_request_id"

type slotRankingEnvelope struct {
	Candidates []slotCandidate `json:"candidates"`
}

type slotCandidate struct {
	SceneID                          string          `json:"scene_id"`
	AnchorStartFrame                 int             `json:"anchor_start_frame"`
	AnchorEndFrame                   int             `json:"anchor_end_frame"`
	QuietWindowSeconds               float64         `json:"quiet_window_seconds"`
	Reasoning                        string          `json:"reasoning"`
	ContextRelevanceScore            float64         `json:"context_relevance_score"`
	NarrativeFitScore                float64         `json:"narrative_fit_score"`
	AnchorContinuityScore            float64         `json:"anchor_continuity_score"`
	QuietWindowScore                 float64         `json:"quiet_window_score"`
	MotionScore                      float64         `json:"motion_score"`
	BoundaryMotionScore              float64         `json:"boundary_motion_score"`
	StartBoundaryMotionScore         float64         `json:"start_boundary_motion_score"`
	EndBoundaryMotionScore           float64         `json:"end_boundary_motion_score"`
	ActionIntensityScore             float64         `json:"action_intensity_score"`
	MaxSubwindowActionIntensity      float64         `json:"max_subwindow_action_intensity"`
	ShotBoundaryDistanceStartSeconds float64         `json:"shot_boundary_distance_start_seconds"`
	ShotBoundaryDistanceEndSeconds   float64         `json:"shot_boundary_distance_end_seconds"`
	StartCutConfidence               float64         `json:"start_cut_confidence"`
	EndCutConfidence                 float64         `json:"end_cut_confidence"`
	StabilityScore                   float64         `json:"stability_score"`
	DialogueActivityScore            float64         `json:"dialogue_activity_score"`
	Metadata                         models.Metadata `json:"metadata"`
}

type slotRankingThresholds struct {
	MotionScoreMax                 float64
	AnchorBoundaryMotionScoreMax   float64
	ActionIntensityScoreMax        float64
	MaxSubwindowActionIntensityMax float64
	ShotBoundaryDistanceSecondsMin float64
	AnchorCutConfidenceMax         float64
	SceneDurationSecondsMin        float64
	QuietWindowSecondsMin          float64
}

func defaultSlotRankingThresholds() slotRankingThresholds {
	return slotRankingThresholds{
		MotionScoreMax:                 0.75,
		AnchorBoundaryMotionScoreMax:   0.85,
		ActionIntensityScoreMax:        0.85,
		MaxSubwindowActionIntensityMax: 0.90,
		ShotBoundaryDistanceSecondsMin: 0.25,
		AnchorCutConfidenceMax:         0.80,
		SceneDurationSecondsMin:        6,
		QuietWindowSecondsMin:          1.5,
	}
}

func (s *JobService) rankSlotsWithOpenAI(ctx context.Context, jobID string, sourceFPS float64, product models.Product, scenes []models.Scene, rejectedSlotIDs []string) ([]models.Slot, string, error) {
	requestPayload, err := buildSlotRankingPrompt(jobID, sourceFPS, product, scenes, rejectedSlotIDs)
	if err != nil {
		return nil, "", fmt.Errorf("build slot ranking prompt: %w", err)
	}

	response, err := s.openAIClient.Complete(ctx, OpenAIRequest{
		JobID:        jobID,
		Purpose:      "phase_2_slot_ranking",
		SystemPrompt: slotRankingSystemPrompt(),
		Prompt:       requestPayload,
		Temperature:  0.1,
	})
	if err != nil {
		return nil, "", err
	}

	candidates, err := parseSlotCandidates(response.Content)
	if err != nil {
		return nil, response.RequestID, err
	}

	slots := buildRankedSlots(jobID, sourceFPS, product, scenes, rejectedSlotIDs, candidates)
	if len(slots) == 0 {
		slots = rankSlots(jobID, sourceFPS, product, scenes, rejectedSlotIDs)
	}
	return slots, response.RequestID, nil
}

func buildSlotRankingPrompt(jobID string, sourceFPS float64, product models.Product, scenes []models.Scene, rejectedSlotIDs []string) (string, error) {
	thresholds := defaultSlotRankingThresholds()
	contentLanguage := detectContentLanguageFromScenes(scenes)
	payload := map[string]any{
		"job_id":            jobID,
		"source_fps":        sourceFPS,
		"content_language":  contentLanguage,
		"product":           product,
		"rejected_slot_ids": rejectedSlotIDs,
		"exclusion_thresholds": map[string]any{
			"motion_score_max":                   thresholds.MotionScoreMax,
			"anchor_boundary_motion_score_max":   thresholds.AnchorBoundaryMotionScoreMax,
			"action_intensity_score_max":         thresholds.ActionIntensityScoreMax,
			"max_subwindow_action_intensity_max": thresholds.MaxSubwindowActionIntensityMax,
			"shot_boundary_distance_seconds_min": thresholds.ShotBoundaryDistanceSecondsMin,
			"anchor_cut_confidence_max":          thresholds.AnchorCutConfidenceMax,
			"scene_duration_seconds_min":         thresholds.SceneDurationSecondsMin,
			"quiet_window_seconds_min":           thresholds.QuietWindowSecondsMin,
		},
		"scenes": scenes,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func slotRankingSystemPrompt() string {
	return strings.Join([]string{
		"You are ranking CAFAI insertion candidates for a source video.",
		"Return strict JSON with a top-level candidates array.",
		"Each candidate must include scene_id, anchor_start_frame, anchor_end_frame, quiet_window_seconds, reasoning, context_relevance_score, narrative_fit_score, anchor_continuity_score, quiet_window_score, motion_score, start_boundary_motion_score, end_boundary_motion_score, action_intensity_score, max_subwindow_action_intensity, shot_boundary_distance_start_seconds, shot_boundary_distance_end_seconds, start_cut_confidence, end_cut_confidence, stability_score, dialogue_activity_score, and metadata.",
		"Propose anchor-frame pairs inside the provided scene bounds.",
		"Physical continuity and low-disruption insertion are the primary goal.",
		"Weak product-context matching must not disqualify a physically valid slot.",
		"Culturally or narratively unexpected product placement is allowed if the insertion remains visually plausible.",
		"Prefer low-motion, quiet scenes and do not repeat previously rejected slot ids.",
		"Keep the candidate reasoning concise and in English.",
		"Do not wrap the JSON in markdown code fences.",
	}, " ")
}

func parseSlotCandidates(content string) ([]slotCandidate, error) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var envelope slotRankingEnvelope
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return nil, fmt.Errorf("decode slot ranking response: %w", err)
	}
	if envelope.Candidates == nil {
		return []slotCandidate{}, nil
	}
	return envelope.Candidates, nil
}

func buildRankedSlots(jobID string, sourceFPS float64, product models.Product, scenes []models.Scene, rejectedSlotIDs []string, candidates []slotCandidate) []models.Slot {
	rejectedSet := make(map[string]struct{}, len(rejectedSlotIDs))
	for _, id := range rejectedSlotIDs {
		rejectedSet[id] = struct{}{}
	}

	sceneByID := make(map[string]models.Scene, len(scenes))
	for _, scene := range scenes {
		sceneByID[scene.ID] = scene
	}

	now := TimestampNow()
	slots := make([]models.Slot, 0, len(candidates))
	for _, candidate := range candidates {
		scene, ok := sceneByID[candidate.SceneID]
		if !ok {
			continue
		}
		slot, valid := buildSlotFromCandidate(jobID, sourceFPS, product, scene, candidate, now)
		if !valid {
			continue
		}
		if _, rejected := rejectedSet[slot.ID]; rejected {
			continue
		}
		slots = append(slots, slot)
	}

	slices.SortFunc(slots, func(left, right models.Slot) int {
		if left.Score > right.Score {
			return -1
		}
		if left.Score < right.Score {
			return 1
		}
		if left.AnchorStartFrame < right.AnchorStartFrame {
			return -1
		}
		if left.AnchorStartFrame > right.AnchorStartFrame {
			return 1
		}
		return strings.Compare(left.ID, right.ID)
	})

	if len(slots) > 3 {
		slots = slots[:3]
	}
	for index := range slots {
		slots[index].Rank = index + 1
	}
	return slots
}

func buildSlotFromCandidate(jobID string, sourceFPS float64, product models.Product, scene models.Scene, candidate slotCandidate, timestamp string) (models.Slot, bool) {
	thresholds := defaultSlotRankingThresholds()
	if sourceFPS <= 0 {
		sourceFPS = 24
	}
	if candidate.AnchorStartFrame < scene.StartFrame || candidate.AnchorEndFrame > scene.EndFrame || candidate.AnchorEndFrame <= candidate.AnchorStartFrame {
		return models.Slot{}, false
	}

	sceneDuration := scene.EndSeconds - scene.StartSeconds
	quietWindowSeconds := firstPositive(candidate.QuietWindowSeconds, floatValue(scene.LongestQuietWindowSeconds, 0))
	motionScore := clamp01(firstPositive(candidate.MotionScore, floatValue(scene.MotionScore, 0)))
	startBoundaryMotion := clamp01(firstPositive(candidate.StartBoundaryMotionScore, candidate.BoundaryMotionScore, motionScore))
	endBoundaryMotion := clamp01(firstPositive(candidate.EndBoundaryMotionScore, candidate.BoundaryMotionScore, motionScore))
	actionIntensity := clamp01(firstPositive(candidate.ActionIntensityScore, floatValue(scene.ActionIntensityScore, motionScore)))
	maxSubwindowAction := clamp01(firstPositive(candidate.MaxSubwindowActionIntensity, actionIntensity))
	startBoundaryDistance := firstPositive(candidate.ShotBoundaryDistanceStartSeconds, float64(candidate.AnchorStartFrame-scene.StartFrame)/sourceFPS)
	endBoundaryDistance := firstPositive(candidate.ShotBoundaryDistanceEndSeconds, float64(scene.EndFrame-candidate.AnchorEndFrame)/sourceFPS)
	startCutConfidence := clamp01(firstPositive(candidate.StartCutConfidence, floatValue(scene.AbruptCutRisk, 0)))
	endCutConfidence := clamp01(firstPositive(candidate.EndCutConfidence, floatValue(scene.AbruptCutRisk, 0)))
	if motionScore > thresholds.MotionScoreMax ||
		startBoundaryMotion > thresholds.AnchorBoundaryMotionScoreMax ||
		endBoundaryMotion > thresholds.AnchorBoundaryMotionScoreMax ||
		actionIntensity > thresholds.ActionIntensityScoreMax ||
		maxSubwindowAction > thresholds.MaxSubwindowActionIntensityMax ||
		startBoundaryDistance <= thresholds.ShotBoundaryDistanceSecondsMin ||
		endBoundaryDistance <= thresholds.ShotBoundaryDistanceSecondsMin ||
		startCutConfidence > thresholds.AnchorCutConfidenceMax ||
		endCutConfidence > thresholds.AnchorCutConfidenceMax ||
		sceneDuration < thresholds.SceneDurationSecondsMin ||
		quietWindowSeconds < thresholds.QuietWindowSecondsMin {
		return models.Slot{}, false
	}

	stabilityScore := clamp01(firstPositive(candidate.StabilityScore, floatValue(scene.StabilityScore, clamp01(1-motionScore))))
	quietWindowScore := clamp01(firstPositive(candidate.QuietWindowScore, quietWindowSeconds/3))
	contextRelevanceScore := clamp01(firstPositive(candidate.ContextRelevanceScore, scoreContextRelevance(product, scene)))
	dialogueScore := clamp01(firstPositive(candidate.DialogueActivityScore, floatValue(scene.DialogueActivityScore, 0)))
	narrativeFitScore := clamp01(firstPositive(candidate.NarrativeFitScore, (1-dialogueScore)*0.50+quietWindowScore*0.30+(1-actionIntensity)*0.20))
	anchorContinuityScore := clamp01(firstPositive(candidate.AnchorContinuityScore, stabilityScore*0.40+(1-maxFloat(startBoundaryMotion, endBoundaryMotion))*0.35+(1-maxFloat(startCutConfidence, endCutConfidence))*0.25))
	slotScore := stabilityScore*0.35 + quietWindowScore*0.25 + narrativeFitScore*0.20 + anchorContinuityScore*0.15 + contextRelevanceScore*0.05

	reasoning := strings.TrimSpace(candidate.Reasoning)
	if reasoning == "" {
		reasoning = fmt.Sprintf(
			"low motion %.2f, %.1f-second quiet window, context match %.2f, stable continuity anchors",
			motionScore,
			quietWindowSeconds,
			contextRelevanceScore,
		)
	}

	slotID := fmt.Sprintf("slot_%s_%s_%d_%d", jobID, scene.ID, candidate.AnchorStartFrame, candidate.AnchorEndFrame)
	metadata := cloneMetadata(candidate.Metadata)
	if metadata == nil {
		metadata = models.Metadata{}
	}
	metadata["quiet_window_score"] = roundScore(quietWindowScore)
	metadata["context_relevance_score"] = roundScore(contextRelevanceScore)
	metadata["narrative_fit_score"] = roundScore(narrativeFitScore)
	metadata["anchor_continuity_score"] = roundScore(anchorContinuityScore)
	metadata["motion_score"] = roundScore(motionScore)
	metadata["start_boundary_motion_score"] = roundScore(startBoundaryMotion)
	metadata["end_boundary_motion_score"] = roundScore(endBoundaryMotion)
	metadata["action_intensity_score"] = roundScore(actionIntensity)
	metadata["max_subwindow_action_intensity"] = roundScore(maxSubwindowAction)
	metadata["shot_boundary_distance_start_seconds"] = roundScore(startBoundaryDistance)
	metadata["shot_boundary_distance_end_seconds"] = roundScore(endBoundaryDistance)
	metadata["start_cut_confidence"] = roundScore(startCutConfidence)
	metadata["end_cut_confidence"] = roundScore(endCutConfidence)

	return models.Slot{
		ID:                    slotID,
		JobID:                 jobID,
		SceneID:               scene.ID,
		AnchorStartFrame:      candidate.AnchorStartFrame,
		AnchorEndFrame:        candidate.AnchorEndFrame,
		QuietWindowSeconds:    roundScore(quietWindowSeconds),
		Score:                 roundScore(slotScore),
		Reasoning:             reasoning,
		Status:                constants.SlotStatusProposed,
		ContextRelevanceScore: floatPtr(roundScore(contextRelevanceScore)),
		NarrativeFitScore:     floatPtr(roundScore(narrativeFitScore)),
		AnchorContinuityScore: floatPtr(roundScore(anchorContinuityScore)),
		Metadata:              metadata,
		CreatedAt:             timestamp,
		UpdatedAt:             timestamp,
	}, true
}

func firstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
