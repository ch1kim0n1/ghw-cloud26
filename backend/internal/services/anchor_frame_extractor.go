package services

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

type FFmpegAnchorFrameExtractor struct {
	artifactsDir string
}

func NewFFmpegAnchorFrameExtractor(artifactsDir string) *FFmpegAnchorFrameExtractor {
	return &FFmpegAnchorFrameExtractor{artifactsDir: artifactsDir}
}

func (e *FFmpegAnchorFrameExtractor) Extract(ctx context.Context, req AnchorFrameRequest) (AnchorFrameArtifacts, error) {
	if err := ensureBinaryOnPath("ffmpeg"); err != nil {
		return AnchorFrameArtifacts{}, err
	}

	sourceFPS := req.SourceFPS
	if sourceFPS <= 0 {
		sourceFPS = 24
	}

	baseDir := filepath.Join(e.artifactsDir, req.JobID, "anchors", req.SlotID)
	startPath := filepath.Join(baseDir, "anchor_start.png")
	endPath := filepath.Join(baseDir, "anchor_end.png")

	if err := extractFrameImage(ctx, req.VideoPath, float64(req.AnchorStartFrame)/sourceFPS, startPath); err != nil {
		return AnchorFrameArtifacts{}, fmt.Errorf("extract start anchor frame: %w", err)
	}
	if err := extractFrameImage(ctx, req.VideoPath, float64(req.AnchorEndFrame)/sourceFPS, endPath); err != nil {
		return AnchorFrameArtifacts{}, fmt.Errorf("extract end anchor frame: %w", err)
	}

	return AnchorFrameArtifacts{
		AnchorStartImagePath: startPath,
		AnchorEndImagePath:   endPath,
	}, nil
}

func extractFrameImage(ctx context.Context, videoPath string, timestampSeconds float64, outputPath string) error {
	if err := ensureParentDirectory(outputPath); err != nil {
		return err
	}

	command := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.3f", timestampSeconds),
		"-frames:v", "1",
		outputPath,
	)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg frame extraction failed: %w: %s", err, string(output))
	}
	return nil
}
