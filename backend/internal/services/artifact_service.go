package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

type PathService struct {
	directories []string
}

func NewPathService(cfg config.Config) *PathService {
	return &PathService{
		directories: []string{
			cfg.UploadProductsDir,
			cfg.UploadCampaignsDir,
			cfg.ArtifactsDir,
			cfg.PreviewsDir,
		},
	}
}

func (s *PathService) EnsureRuntimeDirectories() error {
	for _, dir := range s.directories {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

type LocalStorageService struct{}

func NewLocalStorageService() *LocalStorageService {
	return &LocalStorageService{}
}

func (s *LocalStorageService) Save(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

func (s *LocalStorageService) Read(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return contents, nil
}
