package Crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/rand"
	"crypto/x509"
	"crypto/elliptic"
	dt "GO-DAG/DataTypes"
	"GO-DAG/serialize"
	"encoding/hex"
	"encoding/pem"
	"os"
	"strings"
	"io/ioutil"
	"fmt"
)

func Hash(b []byte) [32]byte {
	h := sha256.Sum256(b)
	return h
}


func PoW(tx *dt.Transaction,difficulty int) {
	for {
		s := serialize.SerializeData(*tx)
		hash := Hash(s)
		h := EncodeToHex(hash[:])
		if h[:difficulty] == strings.Repeat("0",difficulty){
			break
		} 
		tx.Nonce += 1
	}
}

func VerifyPoW(tx dt.Transaction,difficulty int) bool {
	s := serialize.SerializeData(tx)
	hash := Hash(s)
	h := EncodeToHex(hash[:])
	if h[:difficulty] == strings.Repeat("0",difficulty){
		return true
	}
	return false
}

func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

func DecodeToBytes(data string) []byte {
	b,_ := hex.DecodeString(data)
	return b
}

func CheckForKeys() bool {
	// Returns true if file is present in the directory
	filename := "PrivateKey.pem"
	if _,err := os.Stat(filename) ; os.IsNotExist(err) {
		return false
	}
	return true
}

func SerializePublicKey(PublicKey *ecdsa.PublicKey) []byte {
	key := elliptic.Marshal(PublicKey,PublicKey.X,PublicKey.Y)
	return key // uncompressed public key in bytes lenghth 65
}

func SerializePrivateKey(privateKey *ecdsa.PrivateKey) []byte{
	key,_ := x509.MarshalECPrivateKey(privateKey)
	return key
}


func DeserializePublicKey(data []byte) *ecdsa.PublicKey {
	var PublicKey ecdsa.PublicKey
	PublicKey.Curve = elliptic.P256()
	PublicKey.X,PublicKey.Y = elliptic.Unmarshal(elliptic.P256(),data)
	return &PublicKey
}

func DeserializePrivateKey(data []byte) *ecdsa.PrivateKey {
	PrivateKey,_ := x509.ParseECPrivateKey(data)
	return PrivateKey
}


func GenerateKeys() *ecdsa.PrivateKey {
	// function generates keys and creates a pem file with key stored in that
	PrivateKey,err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println(err)
	}
	serialKey := SerializePrivateKey(PrivateKey)
	file,_ := os.OpenFile("PrivateKey.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pemBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: serialKey} // serializing and writing to the pem file
	if err := pem.Encode(file,pemBlock) ; err != nil {
		fmt.Println(err)
	}
	file.Close()
	return PrivateKey
}

func LoadKeys() *ecdsa.PrivateKey {
	// Loads the key from a pem file in directory
	filename := "PrivateKey.pem"
	bytes,err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	block,_ := pem.Decode(bytes)
	PrivateKey := DeserializePrivateKey(block.Bytes)
	return PrivateKey
}

func Sign(hash []byte, key *ecdsa.PrivateKey) []byte {
	// returns the serialized form of the signature
	var signature [72]byte
	if len(hash) != 32 { // check the length of hash
		fmt.Println("Invalid hash")
		return signature[:]
	}
	r,s,_ := ecdsa.Sign(rand.Reader,key,hash)

	copy(signature[:],serialize.PointsToDER(r,s))
	return signature[:]
}

func Verify(signature []byte , PublicKey *ecdsa.PublicKey, hash []byte) bool {

	r,s := serialize.PointsFromDER(signature)

	v := ecdsa.Verify(PublicKey,hash,r,s)
	return v
}