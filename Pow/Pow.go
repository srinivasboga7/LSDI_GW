package Pow

import (
	"strings"
	dt "GO-DAG/datatypes"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
)

//Pow calculates the Nonce field in tx to match the difficulty
func PoW(tx *dt.Transaction,,difficulty int) {
	for {
		s := serialize.SerializeData(*tx)
		hash := Crypto.Hash(s)
		h := Crypto.EncodeToHex(hash[:])
		if h[:difficulty] == strings.Repeat("0",difficulty){
			break
		} 
		tx.Nonce += 1
	}
}

//VerifyPoW verifies if the nonce field of tx matches the difficulty
func VerifyPoW(tx dt.Transaction,difficulty int) bool {
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])
	if h[:difficulty] == strings.Repeat("0",difficulty){
		return true
	}
	return false
}