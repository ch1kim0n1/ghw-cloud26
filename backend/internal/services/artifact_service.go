package services

import (
	"fmt"
	"io"
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
			cfg.CacheDir,
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
	if err := ensureParentDirectory(path); err != nil {
		return err
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

func (s *LocalStorageService) SaveReader(path string, reader io.Reader) error {
	if err := ensureParentDirectory(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("copy file %s: %w", path, err)
	}
	return nil
}

func (s *LocalStorageService) Delete(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file %s: %w", path, err)
	}
	return nil
}

func ensureParentDirectory(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", path, err)
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
