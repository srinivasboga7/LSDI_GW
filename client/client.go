package client

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	pow "GO-DAG/Pow"
	"GO-DAG/consensus"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"crypto/ecdsa"
	"log"
	"net"
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

	copy(tx.LeftTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.001)))
	copy(tx.RightTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.001)))
	pow.PoW(&tx, 3)
	// fmt.Println("After pow")
	b := serialize.Encode32(tx)
	var msg p2p.Msg
	msg.ID = 32
	h := Crypto.Hash(b)
	sign := Crypto.Sign(h[:], cli.PrivateKey)
	msg.Payload = append(b, sign...)
	msg.LenPayload = uint32(len(msg.Payload))
	cli.Send <- msg
	storage.AddTransaction(cli.DAG, tx, sign)
	return
}

// SimulateClient issues fake transactions
func (cli *Client) SimulateClient() {

	if triggerServer() {
		time.Sleep(5 * time.Second)
		i := 0
		for {
			// fmt.Println(i, "===============")
			hash := Crypto.Hash([]byte("Hello,World!"))
			cli.IssueTransaction(hash[:])
			i++
			time.Sleep(400 * time.Millisecond)
		}
	}
}

func triggerServer() bool {
	listener, err := net.Listen("tcp", ":6666")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := listener.Accept()
	buf := make([]byte, 1)
	conn.Read(buf)
	if buf[0] == 0x05 {
		return true
	}
	return false
}
