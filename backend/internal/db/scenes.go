package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type ScenesRepository struct {
	db dbExecutor
}

func NewScenesRepository(db dbExecutor) *ScenesRepository {
	return &ScenesRepository{db: db}
}

func (r *ScenesRepository) Insert(ctx context.Context, scene models.Scene) error {
	keywordsJSON, err := json.Marshal(scene.ContextKeywords)
	if err != nil {
		return fmt.Errorf("marshal scene keywords for %s: %w", scene.ID, err)
	}
	metadataJSON, err := json.Marshal(scene.Metadata)
	if err != nil {
		return fmt.Errorf("marshal scene metadata for %s: %w", scene.ID, err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO scenes (
			id, job_id, scene_number, start_frame, end_frame, start_seconds, end_seconds, motion_score, stability_score, dialogue_activity_score, longest_quiet_window_seconds, narrative_summary, context_keywords_json, action_intensity_score, abrupt_cut_risk, metadata_json, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		scene.ID,
		scene.JobID,
		scene.SceneNumber,
		scene.StartFrame,
		scene.EndFrame,
		scene.StartSeconds,
		scene.EndSeconds,
		scene.MotionScore,
		scene.StabilityScore,
		scene.DialogueActivityScore,
		scene.LongestQuietWindowSeconds,
		nullIfEmpty(scene.NarrativeSummary),
		string(keywordsJSON),
		scene.ActionIntensityScore,
		scene.AbruptCutRisk,
		string(metadataJSON),
		scene.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert scene %s: %w", scene.ID, err)
	}

	return nil
}

func (r *ScenesRepository) ReplaceForJob(ctx context.Context, jobID string, scenes []models.Scene) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM scenes WHERE job_id = ?`, jobID); err != nil {
		return fmt.Errorf("delete scenes for %s: %w", jobID, err)
	}

	for _, scene := range scenes {
		if err := r.Insert(ctx, scene); err != nil {
			return err
		}
	}

	return nil
}

func (r *ScenesRepository) ListByJobID(ctx context.Context, jobID string) ([]models.Scene, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			job_id,
			scene_number,
			start_frame,
			end_frame,
			start_seconds,
			end_seconds,
			motion_score,
			stability_score,
			dialogue_activity_score,
			longest_quiet_window_seconds,
			narrative_summary,
			context_keywords_json,
			action_intensity_score,
			abrupt_cut_risk,
			metadata_json,
			created_at
		FROM scenes
		WHERE job_id = ?
		ORDER BY scene_number ASC
	`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list scenes for %s: %w", jobID, err)
	}
	defer rows.Close()

	scenes := make([]models.Scene, 0)
	for rows.Next() {
		var (
			scene        models.Scene
			narrative    sql.NullString
			contextJSON  string
			metadataJSON sql.NullString
		)
		if err := rows.Scan(
			&scene.ID,
			&scene.JobID,
			&scene.SceneNumber,
			&scene.StartFrame,
			&scene.EndFrame,
			&scene.StartSeconds,
			&scene.EndSeconds,
			&scene.MotionScore,
			&scene.StabilityScore,
			&scene.DialogueActivityScore,
			&scene.LongestQuietWindowSeconds,
			&narrative,
			&contextJSON,
			&scene.ActionIntensityScore,
			&scene.AbruptCutRisk,
			&metadataJSON,
			&scene.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan scene row: %w", err)
		}

		scene.NarrativeSummary = narrative.String
		if contextJSON != "" {
			if err := json.Unmarshal([]byte(contextJSON), &scene.ContextKeywords); err != nil {
				return nil, fmt.Errorf("unmarshal scene keywords %s: %w", scene.ID, err)
			}
		}
		scene.Metadata = models.Metadata{}
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &scene.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal scene metadata %s: %w", scene.ID, err)
			}
		}
		scenes = append(scenes, scene)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scene rows: %w", err)
	}

	return scenes, nil
}
