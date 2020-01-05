package client

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"crypto/ecdsa"
	"time"
)

// Client ...
type Client struct {
	PrivateKey *ecdsa.PrivateKey
	// should I keep the dag here or run a go routine for storage layer
}

// IssueTransaction ...
func (cli *Client) IssueTransaction(hash []byte, send chan p2p.Msg) {
	var tx dt.Transaction
	copy(tx.Hash[:], hash[:])
	tx.Timestamp = time.Now().UnixNano()
	copy(tx.From[:], Crypto.SerializePublicKey(&cli.PrivateKey.PublicKey))
	// tip selection
	// broadcast transaction
	getTips(tx)
	b := serialize.Encode(tx)
	var msg p2p.Msg
	msg.ID = 32
	msg.LenPayload = uint32(len(b))
	msg.Payload = b
	send <- msg
	return
}

// maybe do tip selection on the copy of DAG instead of blocking on a mutex
func getTips(tx dt.Transaction) {

}
