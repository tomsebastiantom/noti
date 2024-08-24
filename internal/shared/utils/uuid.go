package utils

import (
	"github.com/google/uuid"
)

func GenerateUUID() (string) {
	u := uuid.New().String()

	return u
}
