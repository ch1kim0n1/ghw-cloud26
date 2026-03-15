package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const localRenderFallbackMetadataKey = "local_render_fallback_used"

type renderMediaInfo struct {
	Width           int
	Height          int
	DurationSeconds float64
	SourceFPS       float64
	HasAudio        bool
}

func (s *JobService) tryLocalRenderFallback(ctx context.Context, job models.Job, preview models.Preview, slot models.Slot, campaign models.Campaign, artifactManifest models.Metadata, reason string) (bool, error) {
	job.Metadata = ensureJobMetadata(job.Metadata)
	if metadataBool(job.Metadata, localRenderFallbackMetadataKey) {
		return false, nil
	}

	artifactManifest = cloneMetadata(artifactManifest)
	if artifactManifest == nil {
		artifactManifest = models.Metadata{}
	}

	outputPath := filepath.Join(s.previewDir, previewOutputFilename(job.ID))
	renderMetrics, stitchErr := stitchPreviewLocally(ctx, campaign.VideoPath, stringValueOrDefault(slot.GeneratedClipPath), stringValueOrDefault(slot.GeneratedAudioPath), outputPath, slot.AnchorStartFrame, slot.AnchorEndFrame, metadataFloat(job.Metadata, "source_fps"))
	if stitchErr != nil {
		failedPreview := preview
		failedPreview.ArtifactManifest = artifactManifest
		return true, s.failRenderJob(ctx, job, failedPreview, slot.ID, summarizeProviderFailure(reason, stitchErr))
	}

	artifactManifest["local_render_fallback"] = true
	artifactManifest["local_preview_output_path"] = outputPath

	return true, s.persistCompletedLocalRender(ctx, job, preview, slot, campaign, outputPath, artifactManifest, renderMetrics, reason)
}

func stitchPreviewLocally(ctx context.Context, sourceVideoPath, generatedClipPath, generatedAudioPath, outputPath string, anchorStartFrame, anchorEndFrame int, sourceFPS float64) (models.Metadata, error) {
	if err := ensureBinaryOnPath("ffmpeg"); err != nil {
		return nil, err
	}
	if err := ensureBinaryOnPath("ffprobe"); err != nil {
		return nil, err
	}
	if err := ensureParentDirectory(outputPath); err != nil {
		return nil, err
	}

	sourceInfo, err := inspectRenderMedia(ctx, sourceVideoPath)
	if err != nil {
		return nil, fmt.Errorf("inspect source video: %w", err)
	}
	clipInfo, err := inspectRenderMedia(ctx, generatedClipPath)
	if err != nil {
		return nil, fmt.Errorf("inspect generated clip: %w", err)
	}
	if !sourceInfo.HasAudio {
		return nil, fmt.Errorf("source video is missing audio; local render fallback currently requires source audio")
	}

	audioSourcePath := strings.TrimSpace(generatedAudioPath)
	if audioSourcePath == "" && clipInfo.HasAudio {
		audioSourcePath = generatedClipPath
	}
	useGeneratedSilence := audioSourcePath == ""

	startSeconds := float64(anchorStartFrame) / sourceFPS
	endSeconds := float64(anchorEndFrame) / sourceFPS
	if endSeconds < startSeconds {
		return nil, fmt.Errorf("invalid anchor frames: start=%d end=%d", anchorStartFrame, anchorEndFrame)
	}

	args := []string{
		"-y",
		"-i", sourceVideoPath,
		"-i", generatedClipPath,
	}
	audioInputIndex := 1
	if audioSourcePath != "" && audioSourcePath != generatedClipPath {
		args = append(args, "-i", audioSourcePath)
		audioInputIndex = 2
	}
	if useGeneratedSilence {
		args = append(args, "-f", "lavfi", "-i", fmt.Sprintf("anullsrc=r=44100:cl=stereo:d=%.3f", clipInfo.DurationSeconds))
		audioInputIndex = len(args)/2 - 1
	}

	filter := fmt.Sprintf(
		"[0:v]trim=0:%.3f,setpts=PTS-STARTPTS[v0];"+
			"[0:a]atrim=0:%.3f,asetpts=PTS-STARTPTS[a0];"+
			"[1:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,fps=%.12f,format=yuv420p,setsar=1,setpts=PTS-STARTPTS[v1];"+
			"[%d:a]aresample=44100,asetpts=PTS-STARTPTS[a1];"+
			"[0:v]trim=start=%.3f,setpts=PTS-STARTPTS[v2];"+
			"[0:a]atrim=start=%.3f,asetpts=PTS-STARTPTS[a2];"+
			"[v0][a0][v1][a1][v2][a2]concat=n=3:v=1:a=1[vout][aout]",
		startSeconds,
		startSeconds,
		sourceInfo.Width,
		sourceInfo.Height,
		sourceInfo.Width,
		sourceInfo.Height,
		sourceInfo.SourceFPS,
		audioInputIndex,
		endSeconds,
		endSeconds,
	)

	args = append(args,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", "[aout]",
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "18",
		"-c:a", "aac",
		"-b:a", "192k",
		outputPath,
	)

	command := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg local stitch failed: %w: %s", err, string(output))
	}

	outputInfo, err := inspectRenderMedia(ctx, outputPath)
	if err != nil {
		return nil, fmt.Errorf("inspect local render output: %w", err)
	}

	insertedDuration := outputInfo.DurationSeconds - sourceInfo.DurationSeconds
	if insertedDuration <= 0 {
		insertedDuration = clipInfo.DurationSeconds
	}

	return models.Metadata{
		"render_provider_used":            "local_ffmpeg_fallback",
		"source_duration_seconds":         roundScore(sourceInfo.DurationSeconds),
		"preview_duration_seconds":        roundScore(outputInfo.DurationSeconds),
		"inserted_duration_seconds":       roundScore(insertedDuration),
		"generated_clip_duration_seconds": roundScore(clipInfo.DurationSeconds),
		"anchor_start_frame":              anchorStartFrame,
		"anchor_end_frame":                anchorEndFrame,
	}, nil
}

func (s *JobService) persistCompletedLocalRender(ctx context.Context, job models.Job, preview models.Preview, slot models.Slot, campaign models.Campaign, outputPath string, artifactManifest models.Metadata, renderMetrics models.Metadata, reason string) error {
	completedAt := TimestampNow()
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	jobRepo := db.NewJobsRepository(tx)
	previewRepo := db.NewPreviewsRepository(tx)
	logRepo := db.NewJobLogsRepository(tx)

	current, err := jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		return err
	}
	current.Metadata = ensureJobMetadata(current.Metadata)
	delete(current.Metadata, internalRenderRequestIDKey)
	delete(current.Metadata, internalRenderPayloadRef)
	current.Metadata[localRenderFallbackMetadataKey] = true
	current.Metadata["render_provider_used"] = "local_ffmpeg_fallback"
	if strings.TrimSpace(reason) != "" {
		current.Metadata["render_fallback_reason"] = reason
	}
	current.Status = constants.JobStatusCompleted
	current.CurrentStage = constants.StageRenderPoll
	current.ProgressPercent = 100
	current.ErrorCode = nil
	current.ErrorMessage = nil
	current.CompletedAt = &completedAt
	if err := jobRepo.UpdateState(ctx, current); err != nil {
		return err
	}

	previewDuration := metadataFloat(renderMetrics, "preview_duration_seconds")
	if err := previewRepo.MarkCompleted(ctx, models.Preview{
		ID:               preview.ID,
		JobID:            job.ID,
		SlotID:           slot.ID,
		Status:           "completed",
		OutputVideoPath:  outputPath,
		DurationSeconds:  previewDuration,
		RenderRetryCount: preview.RenderRetryCount,
		CreatedAt:        preview.CreatedAt,
		CompletedAt:      &completedAt,
		ArtifactManifest: artifactManifest,
		RenderMetrics:    renderMetrics,
	}); err != nil {
		return err
	}

	if err := logRepo.Insert(ctx, models.JobLog{
		JobID:     current.ID,
		Timestamp: completedAt,
		EventType: "stage_completed",
		StageName: constants.StageRenderPoll,
		Message:   "preview render complete via local fallback",
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func inspectRenderMedia(ctx context.Context, path string) (renderMediaInfo, error) {
	command := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	output, err := command.Output()
	if err != nil {
		return renderMediaInfo{}, err
	}

	var payload struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType    string `json:"codec_type"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			AvgFrameRate string `json:"avg_frame_rate"`
			RFrameRate   string `json:"r_frame_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return renderMediaInfo{}, err
	}

	info := renderMediaInfo{}
	if payload.Format.Duration != "" {
		duration, err := strconv.ParseFloat(payload.Format.Duration, 64)
		if err != nil {
			return renderMediaInfo{}, err
		}
		info.DurationSeconds = duration
	}

	for _, stream := range payload.Streams {
		switch stream.CodecType {
		case "video":
			info.Width = stream.Width
			info.Height = stream.Height
			fps, err := parseFrameRate(stream.AvgFrameRate)
			if err != nil || fps <= 0 {
				fps, err = parseFrameRate(stream.RFrameRate)
				if err != nil {
					return renderMediaInfo{}, err
				}
			}
			info.SourceFPS = fps
		case "audio":
			info.HasAudio = true
		}
	}

	if info.Width <= 0 || info.Height <= 0 || info.SourceFPS <= 0 {
		return renderMediaInfo{}, fmt.Errorf("missing render media stream metadata for %s", path)
	}
	return info, nil
}
