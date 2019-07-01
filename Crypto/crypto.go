package Crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/rand"
	"crypto/x509"
	"crypto/elliptic"
	"encoding/pem"
	"os"
	"ioutil"
)

func hash(b []byte) [32]byte {
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
	key,_ := MarshalECPrivateKey(privateKey)
	return key
}

func DeserializePublicKey(data []byte) *ecdsa.PublicKey {
	PublicKey,_ := x509.ParsePKIXPublicKey(data)
	return PublicKey
}

func DeserializePrivateKey(data []byte) *ecdsa.PrivateKey {
	PrivateKey,_ := x509.ParseECPrivateKey(data)
	return PrivateKey
}

func SerializeSignature(r *big.Int,s *big.Int) []byte{

}

func DeserializeSignature(sig []byte) *big.Int {

}

func GenerateKeys() *ecdsa.PrivateKey {
	// function generates keys and creates a pem file with key stored in that
	PrivateKey,err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println(err)
	}
	serialKey := SerializePrivateKey(&PrivateKey)
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
	block,rest := pem.Decode(bytes)
	PrivateKey := DeserializePrivateKey(block.Bytes)
	return PrivateKey
}

func Sign(hash []byte, key *ecdsa.PrivateKey) []byte {
	// returns the serialized form of the signature
	// figure out how to serialize signature
	var signature []byte
	if len(hash) != 32 { // check the length of hash
		fmt.Println("Invalid hash")
		return signature
	}
	r,s,err := ecdsa.Sign(rand.Reader,key,hash)

}