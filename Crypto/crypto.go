package Crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/rand"
	"crypto/x509"
	"crypto/elliptic"
	"encoding/pem"
	"math/big"
	"os"
	"io/ioutil"
	"fmt"
)

func Hash(b []byte) [32]byte {
	h := sha256.Sum256(b)
	return h
}

func PoW() {
			
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
	key,_ := x509.MarshalPKIXPublicKey(PublicKey)
	return key // uncompressed public key in bytes
}

func SerializePrivateKey(privateKey *ecdsa.PrivateKey) []byte{
	key,_ := x509.MarshalECPrivateKey(privateKey)
	return key
}


func DeserializePublicKey(data []byte) *ecdsa.PublicKey {
	PublicKey,_ := x509.ParsePKIXPublicKey(data)
	PubKey,_ := PublicKey.(*ecdsa.PublicKey)
	return PubKey
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
	var signature []byte
	if len(hash) != 32 { // check the length of hash
		fmt.Println("Invalid hash")
		return signature
	}
	r,s,_ := ecdsa.Sign(rand.Reader,key,hash)

	signature = r.Bytes()
	signature = append(signature,s.Bytes()...)
	return signature
}

func Verify(signature []byte , PublicKey *ecdsa.PublicKey, hash []byte) bool {
	if len(signature) != 64 {
		fmt.Println("Invalid Signature")
		return false
	}
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(signature[:32])
	s.SetBytes(signature[32:])

	return ecdsa.Verify(PublicKey,hash,r,s)
}