package services

import (
	"fmt"

	"github.com/google/uuid"
)

func NewPrefixedID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewString())
}
