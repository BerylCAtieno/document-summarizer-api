package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random hex ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
