package main

import (
	"LSDI_GW/Crypto"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const GWAddress = "172.18.0.6:8989"

// uploads the data to the blockchain
func uploadToBlockchain(hash string) string {
	url := "http://" + GWAddress + "/api?q="

	response, err := http.Get(url + hash)
	if err != nil {
		log.Fatal(err)
	}

	TxID, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	txid := hex.EncodeToString(TxID)

	return txid
}

func main() {

	for i := 0; i < 15; i++ {
		hash := Crypto.Hash([]byte(time.Now().String()))
		txID := uploadToBlockchain(Crypto.EncodeToHex(hash[:]))
		fmt.Println(txID)
		time.Sleep(time.Second)
	}
}
