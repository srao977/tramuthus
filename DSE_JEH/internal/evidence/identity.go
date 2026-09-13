package evidence

import (
	"crypto/sha256"
	"encoding/hex"
)

func ID(parts ...string) string {
	hash := sha256.New()
	for index, part := range parts {
		if index > 0 {
			_, _ = hash.Write([]byte{0})
		}
		_, _ = hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
