package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type SlotsRepository struct {
	db dbExecutor
}

func NewSlotsRepository(db dbExecutor) *SlotsRepository {
	return &SlotsRepository{db: db}
}

func (r *SlotsRepository) ReplaceForJob(ctx context.Context, jobID string, slots []models.Slot) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM slots WHERE job_id = ?`, jobID); err != nil {
		return fmt.Errorf("delete slots for %s: %w", jobID, err)
	}

	for _, slot := range slots {
		metadataJSON, err := json.Marshal(slot.Metadata)
		if err != nil {
			return fmt.Errorf("marshal slot metadata for %s: %w", slot.ID, err)
		}

		_, err = r.db.ExecContext(ctx, `
			INSERT INTO slots (
				id, job_id, scene_id, rank, anchor_start_frame, anchor_end_frame, quiet_window_seconds, score, context_relevance_score, narrative_fit_score, anchor_continuity_score, reasoning, status, suggested_product_line, final_product_line, product_line_mode, rejection_note, generated_clip_path, generated_audio_path, generation_error, metadata_json, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			slot.ID,
			slot.JobID,
			slot.SceneID,
			slot.Rank,
			slot.AnchorStartFrame,
			slot.AnchorEndFrame,
			slot.QuietWindowSeconds,
			slot.Score,
			slot.ContextRelevanceScore,
			slot.NarrativeFitScore,
			slot.AnchorContinuityScore,
			slot.Reasoning,
			slot.Status,
			slot.SuggestedProductLine,
			slot.FinalProductLine,
			slot.ProductLineMode,
			rejectionNoteFromMetadata(slot.Metadata),
			slot.GeneratedClipPath,
			slot.GeneratedAudioPath,
			slot.GenerationError,
			string(metadataJSON),
			slot.CreatedAt,
			slot.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert slot %s: %w", slot.ID, err)
		}
	}

	return nil
}

func (r *SlotsRepository) ListByJobID(ctx context.Context, jobID string) ([]models.Slot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			job_id,
			scene_id,
			rank,
			anchor_start_frame,
			anchor_end_frame,
			quiet_window_seconds,
			score,
			context_relevance_score,
			narrative_fit_score,
			anchor_continuity_score,
			reasoning,
			status,
			suggested_product_line,
			final_product_line,
			product_line_mode,
			rejection_note,
			generated_clip_path,
			generated_audio_path,
			generation_error,
			metadata_json,
			created_at,
			updated_at
		FROM slots
		WHERE job_id = ?
		ORDER BY CASE status WHEN 'proposed' THEN 0 WHEN 'selected' THEN 1 WHEN 'rejected' THEN 2 ELSE 3 END, rank ASC, datetime(updated_at) DESC, id ASC
	`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list slots for %s: %w", jobID, err)
	}
	defer rows.Close()

	slots := make([]models.Slot, 0)
	for rows.Next() {
		slot, scanErr := scanSlot(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan slot row: %w", scanErr)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate slot rows: %w", err)
	}

	return slots, nil
}

func (r *SlotsRepository) GetByID(ctx context.Context, jobID, slotID string) (models.Slot, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			job_id,
			scene_id,
			rank,
			anchor_start_frame,
			anchor_end_frame,
			quiet_window_seconds,
			score,
			context_relevance_score,
			narrative_fit_score,
			anchor_continuity_score,
			reasoning,
			status,
			suggested_product_line,
			final_product_line,
			product_line_mode,
			rejection_note,
			generated_clip_path,
			generated_audio_path,
			generation_error,
			metadata_json,
			created_at,
			updated_at
		FROM slots
		WHERE job_id = ? AND id = ?
		LIMIT 1
	`, jobID, slotID)

	slot, err := scanSlot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Slot{}, ErrNotFound
	}
	if err != nil {
		return models.Slot{}, fmt.Errorf("query slot %s for job %s: %w", slotID, jobID, err)
	}

	return slot, nil
}

func (r *SlotsRepository) UpdateRejected(ctx context.Context, jobID, slotID, note, updatedAt string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE slots
		SET status = 'rejected', rejection_note = ?, updated_at = ?
		WHERE job_id = ? AND id = ?
	`, nullIfEmpty(note), updatedAt, jobID, slotID)
	if err != nil {
		return fmt.Errorf("reject slot %s for job %s: %w", slotID, jobID, err)
	}
	return nil
}

type slotScanner interface {
	Scan(dest ...any) error
}

func scanSlot(scanner slotScanner) (models.Slot, error) {
	var (
		slot                 models.Slot
		suggestedProductLine sql.NullString
		finalProductLine     sql.NullString
		productLineMode      sql.NullString
		rejectionNote        sql.NullString
		generatedClipPath    sql.NullString
		generatedAudioPath   sql.NullString
		generationError      sql.NullString
		metadataJSON         sql.NullString
	)

	err := scanner.Scan(
		&slot.ID,
		&slot.JobID,
		&slot.SceneID,
		&slot.Rank,
		&slot.AnchorStartFrame,
		&slot.AnchorEndFrame,
		&slot.QuietWindowSeconds,
		&slot.Score,
		&slot.ContextRelevanceScore,
		&slot.NarrativeFitScore,
		&slot.AnchorContinuityScore,
		&slot.Reasoning,
		&slot.Status,
		&suggestedProductLine,
		&finalProductLine,
		&productLineMode,
		&rejectionNote,
		&generatedClipPath,
		&generatedAudioPath,
		&generationError,
		&metadataJSON,
		&slot.CreatedAt,
		&slot.UpdatedAt,
	)
	if err != nil {
		return models.Slot{}, err
	}

	if suggestedProductLine.Valid {
		value := suggestedProductLine.String
		slot.SuggestedProductLine = &value
	}
	if finalProductLine.Valid {
		value := finalProductLine.String
		slot.FinalProductLine = &value
	}
	if productLineMode.Valid {
		value := productLineMode.String
		slot.ProductLineMode = &value
	}
	if generatedClipPath.Valid {
		value := generatedClipPath.String
		slot.GeneratedClipPath = &value
	}
	if generatedAudioPath.Valid {
		value := generatedAudioPath.String
		slot.GeneratedAudioPath = &value
	}
	if generationError.Valid {
		value := generationError.String
		slot.GenerationError = &value
	}
	slot.Metadata = models.Metadata{}
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &slot.Metadata); err != nil {
			return models.Slot{}, fmt.Errorf("unmarshal slot metadata %s: %w", slot.ID, err)
		}
	}
	if rejectionNote.Valid {
		if slot.Metadata == nil {
			slot.Metadata = models.Metadata{}
		}
		slot.Metadata["rejection_note"] = rejectionNote.String
	}

	return slot, nil
}

func rejectionNoteFromMetadata(metadata models.Metadata) any {
	if metadata == nil {
		return nil
	}
	if value, ok := metadata["rejection_note"].(string); ok && value != "" {
		return value
	}
	return nil
}
