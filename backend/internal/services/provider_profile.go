package services

import (
	"fmt"
	"strings"
)

const (
	ProviderProfileAzure = "azure"
	ProviderProfileVultr = "vultr"
)

func normalizeProviderProfile(profile string) string {
	trimmed := strings.ToLower(strings.TrimSpace(profile))
	if trimmed == "" {
		return ProviderProfileAzure
	}
	return trimmed
}

func validateProviderProfile(profile string) error {
	switch normalizeProviderProfile(profile) {
	case ProviderProfileAzure, ProviderProfileVultr:
		return nil
	default:
		return fmt.Errorf("unsupported provider profile %q; expected azure or vultr", profile)
	}
}
