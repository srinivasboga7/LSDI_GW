package client

import (
	"encoding/json"
	dt "GO-DAG/DataTypes"
	"time"
	"net/http"
	"github.com/google/uuid"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/consensus"
	"GO-DAG/storage"
	"crypto/ecdsa"
	"fmt"
	"bytes"
)

type sensordata struct {
	sensorName string
	sensorID string
	data string
	start int64
	SmID string
}

type postRequest struct {
	data sensordata
	ID [16]byte
}

// GenerateMessage creates a byte slice to be sent over a network
// for a particular transaction.
func GenerateMessage(b []byte, signature []byte) []byte {
	var l uint32
	l = uint32(len(b))
	var magicNumber uint32
	magicNumber = 1
	payload := append(b,signature...)
	payload = append(serialize.EncodeToBytes(l),payload...)
	payload = append(serialize.EncodeToBytes(magicNumber),payload...)
	return payload
}

// BroadcastTransaction sends the transaction to all the peers
func BroadcastTransaction(b []byte, p *dt.Peers) {
	p.Mux.Lock()
	for _,conn := range p.Fds {
		conn.Write(b)
	}
	p.Mux.Unlock()
}

// GenerateSignature uses Crypto pkg to genearte a signature.
// It is a wrapper aroung Crypto.Sign function
func GenerateSignature(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
	return signature
}

// Copy generates the Deep copy of the DAG
func Copy(dag *dt.DAG) *dt.DAG {
	bytes,_ := json.Marshal(dag.Graph)
	var copyGraph map[string] dt.Vertex
	json.Unmarshal(bytes,&copyGraph)
	var copyDag dt.DAG
	copyDag.Graph = copyGraph
	copyDag.Genisis = dag.Genisis
	return &copyDag
}

func fakeSensorData(data *sensordata) {
	data.start = time.Now().Unix()
}

// SimulateClient is used for testing by sending fake data as transactions.
func SimulateClient(p *dt.Peers, PrivateKey *ecdsa.PrivateKey, dag *dt.DAG, url string) {

	var tx dt.Transaction
	var fakeData sensordata
	fakeData.sensorName = "livingroom-sensor"
	ID := Crypto.Hash(Crypto.SerializePublicKey(&PrivateKey.PublicKey))
	fakeData.sensorID = string(ID[:])
	fakeData.SmID = "29042884"
	fakeData.data = "lighton"

	for {
		p.Mux.Lock()
		l := len(p.Fds)
		p.Mux.Unlock()
		if l > 0 {
			break
		}
		time.Sleep(2*time.Second)
	}

	fmt.Println("Generating Transactions")

	for {
		fakeSensorData(&fakeData)
		tx.Timestamp = time.Now().Unix()
		serial,_ := json.Marshal(fakeData)
		tx.Hash = Crypto.Hash(serial)
		tx.TxID = uuid.New()
		var request postRequest 
		request.data = fakeData
		request.ID = tx.TxID
		b := bytes.NewReader(serial)
		http.Post(url,"application/json",b)
		copy(tx.From[:],Crypto.SerializePublicKey(&PrivateKey.PublicKey))
		dag.Mux.Lock()
		copyDag := Copy(dag)
		dag.Mux.Unlock()
		copy(tx.LeftTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
		copy(tx.RightTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
		Crypto.PoW(&tx,2)
		buffer := serialize.SerializeData(tx)
		sign := GenerateSignature(buffer,PrivateKey)
		msg := GenerateMessage(buffer,sign)
		storage.AddTransaction(dag,tx,sign)
		BroadcastTransaction(msg,p)
		time.Sleep(5*time.Second)
	}
}
