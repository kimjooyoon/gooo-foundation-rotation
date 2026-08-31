package protocol

import (
	"crypto/sha256"
	"encoding/hex"
)

func DigestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func Digest(value any) (string, error) {
	raw, err := CanonicalJSON(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(raw), nil
}
