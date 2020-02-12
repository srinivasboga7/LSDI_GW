package sharding

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	pow "GO-DAG/Pow"
	"GO-DAG/serialize"
	"errors"
	"fmt"
	"log"
	"time"
)

//VerifyDiscovery verifies shard signal from discovery
func VerifyDiscovery(PuK, msg dt.ShardSignal, signature []byte) bool {
	//Deserialise the message
	//Verify signature using PuK
	//Return valid or not
	return true
}

//VerifyShardTransaction verifies shard transaction
func VerifyShardTransaction(tx dt.ShardTransaction, signature []byte, difficulty int) bool {
	//Verify Puk
	//Verify signature
	//Verify Pow
	//Verify Shard number
	s := serialize.Encode(tx)
	SerialKey := tx.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature, PublicKey, h[:])
	if sigVerify == false {
		log.Println("INVALID SIGNATURE")
		fmt.Println()
	}
	return sigVerify && pow.VerifyPoW(tx, difficulty) && tx.ShardNo == tx.Nonce%2
}

//MakeShardingtx Call on recieving sharding signal from discovery after forwarding it
func MakeShardingtx(Puk, Signal dt.ShardSignal, signature []byte) (dt.ShardTransaction, error) {
	difficulty := 4
	if VerifyDiscovery(PukDiscovery, Signal, signature) {
		//Create transaction
		var tx dt.ShardTransaction
		tx.From = Puk
		tx.Timestamp = time.Now().UnixNano()
		tx.Nonce = 0
		tx.ShardNo = 0
		//Do PoW
		pow.PoW(&tx, difficulty)
		//Wait for recieving messages
		//	//Verify each recvd pow and add to buffer
		//	//Wait till threshold or timeout
		//Update peers
		return tx, nil
	} else {
		return nil, errors.New("Invalid Shard Signal")
	}
}
