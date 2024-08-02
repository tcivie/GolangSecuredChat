package utils

import (
	"crypto/rsa"
	"fmt"
)

func DebugPrintPublicKey(key *rsa.PublicKey) string {
	return fmt.Sprintf("N: %x... (len: %d bits), E: %d", key.N.Bytes()[:20], key.N.BitLen(), key.E)
}
