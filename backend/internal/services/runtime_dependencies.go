package services

import (
	"fmt"
	"os/exec"
	"runtime"
)

func EnsureRuntimeDependencies() error {
	if err := ensureBinaryOnPath("ffprobe"); err != nil {
		return err
	}
	if err := ensureBinaryOnPath("ffmpeg"); err != nil {
		return err
	}
	return nil
}

func ensureBinaryOnPath(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required on PATH: %s", name, binaryInstallHint(name))
	}
	return nil
}

func binaryInstallHint(name string) string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("install %s with Homebrew, for example `brew install ffmpeg`", name)
	case "windows":
		return fmt.Sprintf("install %s and make sure the binary directory is on PATH", name)
	default:
		return fmt.Sprintf("install %s and make sure it is available on PATH", name)
	}
}
