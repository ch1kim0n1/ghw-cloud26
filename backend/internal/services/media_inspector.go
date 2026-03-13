package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type MediaInfo struct {
	FormatName      string
	VideoCodec      string
	DurationSeconds float64
	SourceFPS       float64
}

type MediaInspector struct{}

func NewMediaInspector() *MediaInspector {
	return &MediaInspector{}
}

func (i *MediaInspector) Inspect(ctx context.Context, path string) (MediaInfo, error) {
	command := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := command.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return MediaInfo{}, fmt.Errorf("run ffprobe on %s: %w", path, ensureBinaryOnPath("ffprobe"))
		}
		return MediaInfo{}, fmt.Errorf("run ffprobe on %s: %w", path, err)
	}

	var payload struct {
		Format struct {
			FormatName string `json:"format_name"`
			Duration   string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType    string `json:"codec_type"`
			CodecName    string `json:"codec_name"`
			AvgFrameRate string `json:"avg_frame_rate"`
			RFrameRate   string `json:"r_frame_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &payload); err != nil {
		return MediaInfo{}, fmt.Errorf("decode ffprobe output for %s: %w", path, err)
	}

	durationSeconds, err := strconv.ParseFloat(payload.Format.Duration, 64)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("parse duration for %s: %w", path, err)
	}

	info := MediaInfo{
		FormatName:      payload.Format.FormatName,
		DurationSeconds: durationSeconds,
	}

	for _, stream := range payload.Streams {
		if stream.CodecType != "video" {
			continue
		}
		info.VideoCodec = stream.CodecName
		fps, err := parseFrameRate(stream.AvgFrameRate)
		if err != nil || fps <= 0 {
			fps, err = parseFrameRate(stream.RFrameRate)
			if err != nil {
				return MediaInfo{}, fmt.Errorf("parse frame rate for %s: %w", path, err)
			}
		}
		info.SourceFPS = fps
		break
	}

	if info.VideoCodec == "" || info.SourceFPS <= 0 {
		return MediaInfo{}, fmt.Errorf("missing video stream metadata for %s", path)
	}

	return info, nil
}

func (i *MediaInspector) ValidatePhaseOneVideo(info MediaInfo) error {
	if !strings.Contains(info.FormatName, "mp4") || info.VideoCodec != "h264" {
		return fmt.Errorf("unsupported video format: format=%s codec=%s", info.FormatName, info.VideoCodec)
	}
	if info.DurationSeconds < 600 || info.DurationSeconds > 1200 {
		return fmt.Errorf("unsupported video duration: %.3f", info.DurationSeconds)
	}
	return nil
}

func parseFrameRate(value string) (float64, error) {
	if value == "" || value == "0/0" {
		return 0, fmt.Errorf("empty frame rate")
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return strconv.ParseFloat(value, 64)
	}
	numerator, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	denominator, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	if denominator == 0 {
		return 0, fmt.Errorf("invalid frame rate denominator")
	}
	return numerator / denominator, nil
}
