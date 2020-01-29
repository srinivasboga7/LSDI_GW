package client

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/consensus"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"crypto/ecdsa"
	"log"
	"time"
)

// Client ...
type Client struct {
	PrivateKey *ecdsa.PrivateKey
	// should I keep the dag here or run a go routine for storage layer
	Send chan p2p.Msg
	DAG  *dt.DAG
}

// IssueTransaction ...
func (cli *Client) IssueTransaction(hash []byte) {
	var tx dt.Transaction
	copy(tx.Hash[:], hash[:])
	tx.Timestamp = time.Now().UnixNano()
	copy(tx.From[:], Crypto.SerializePublicKey(&cli.PrivateKey.PublicKey))
	// tip selection
	// broadcast transaction
	copy(tx.LeftTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.01)))
	copy(tx.RightTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.01)))
	b := serialize.Encode(tx)
	var msg p2p.Msg
	msg.ID = 32
	h := Crypto.Hash(b)
	sign := Crypto.Sign(h[:], cli.PrivateKey)
	msg.Payload = append(b, sign...)
	msg.LenPayload = uint32(len(msg.Payload))
	log.Println(Crypto.EncodeToHex(tx.LeftTip[:]))
	cli.Send <- msg
	storage.AddTransaction(cli.DAG, tx, sign)
	return
}

// SimulateClient issues fake transactions
func (cli *Client) SimulateClient() {
	for {
		time.Sleep(2 * time.Second)
		hash := Crypto.Hash([]byte("Hello,World!"))
		cli.IssueTransaction(hash[:])
	}
}
