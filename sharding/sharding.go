package sharding

import (
	dt "GO-DAG/DataTypes"
	pow "GO-DAG/Pow"
	"time"
	"math/rand"
)

func VerifyDiscovery(PuK, msg dt.ShardSignal, signature []byte) bool {
	//Deserialise the message
	//Verify signature using PuK
	//Return valid or not
}


// DeserialiseMsg deserialises the message stream shardsignal
func DeserialiseMsg(stream []byte) dt.ShardSignal {
	//Deserialise to shardsignal format
	return 
}

func Broadcast()

func StartServer(myShardNo int, difficulty int) []string {
	var ShardNodesIpList []string
	ShardNodesIpList = StartServer(tx[ShardNo], difficulty)
	rand.Seed(time.Now().UnixNano())
	orderToConnect := rand.Perm(len(ShardNodesIpList))
	DisconnectOldConn()
	MakeNewConn(orderToConnect)
	var ShardNodes []string
	while(threshold not reached) {
		//Recieve transaction
		if rctx[ShardNo]==myShardNo {
			if VerifyShardTransaction(rctx, signature, difficulty) {
				ShardNodes = append(ShardNodes,ip)
			}
		}
	}
}


func VerifyShardTransaction(tx dt.ShardTransaction, signature []byte, difficulty int) bool {
	//Verify Puk
	//Verify signature
	//Verify Pow
	//Verify Shard number
	s := serialize.SerializeData(tx)
	SerialKey := tx.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature,PublicKey,h[:])
	if sigVerify == false {
		log.Println("INVALID SIGNATURE")
		fmt.Println()
	}
	return sigVerify && pow.VerifyPoW(tx,difficulty) && tx[ShardNo] == tx[Nonce]%2
}

//RecvdMssg Call on recieving sharding signal from discovery after forwarding it
func StartSharding(Puk, msgStream []byte, signature []byte) (dt.ShardTransaction,error) {
	//Deserialise message, seperate signature
	//Verify if recieved from valid gateway
	difficulty := 4
	signal, sign := serialize.DeserializeShardSignal(msgStream, "ShardSignal")
	if VerifyDiscovery(PukDiscovery, signal, sign) {
		//Create transaction
		var tx dt.ShardTransaction
		tx.From = Puk
		tx.Timestamp = time.Now().UnixNano()
		tx.Nonce = 0
		tx.ShardNo = 0
		//Do PoW
		pow.PoW(&tx, difficulty)
		//Broadcast to network
		Broadcast()
		//Wait for recieving messages
		//	//Verify each recvd pow and add to buffer
		//	//Wait till threshold or timeout
		//Update peers
		return tx, nil
	} else {
		return nil, errors.New("Invalid Shard Signal")
	}
}