package Crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	dt "GO-DAG/DataTypes"
)

func hash(b []byte) [32]byte {
	h := sha256.Sum256(b)
	return h
}

