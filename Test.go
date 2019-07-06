package main


import(
	"fmt"
	"GO-DAG/Crypto"
)

func main() {
	// test
	PrivateKey := Crypto.GenerateKeys()
	message := "I Love You 3000"
	h := Crypto.Hash([]byte(message))
	sign := Crypto.Sign(h[:],PrivateKey)
	SerialKey := Crypto.SerializePublicKey(&PrivateKey.PublicKey)
	fmt.Println(len(SerialKey))
	PublicKey := Crypto.DeserializePublicKey(SerialKey)
	v := Crypto.Verify(sign,PublicKey,h[:])
	fmt.Println(v)
	//
}