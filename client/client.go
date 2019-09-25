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
	"net"
)

type Client struct {
	PrivateKey *ecdsa.PrivateKey
	Dag *dt.DAG
	Peers *dt.Peers
}

type sensordata struct {
	Data string
	SensorName string
	//SensorID string
	SmID string
	Start int64
	//End int64
}

type postRequest struct {
	serialData []byte
	TxID string
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


/*
// SimulateClient is used for testing by sending fake data as transactions.
func SimulateClient(p *dt.Peers, PrivateKey *ecdsa.PrivateKey, dag *dt.DAG, url string) {

	var tx dt.Transaction

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
		s,_ := json.Marshal(fakeData)
		tx.Hash = Crypto.Hash(s)
		tx.TxID = uuid.New()
		var request postRequest 
		request.Meter = fakeData
		request.TxID = Crypto.EncodeToHex(tx.TxID[:])
		serial,_ := json.Marshal(request)
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
		time.Sleep(time.Second)
	}
}
*/

func (cli *Client)createTransaction(data []byte) []byte{
	privateKey := cli.PrivateKey
	var tx dt.Transaction
	tx.Timestamp = time.Now().Unix()
	tx.Hash = Crypto.Hash(data)
	tx.TxID = uuid.New()
	var request postRequest 
	request.serialData = data
	request.TxID = Crypto.EncodeToHex(tx.TxID[:])
	serial,_ := json.Marshal(request)
	copy(tx.From[:],Crypto.SerializePublicKey(&privateKey.PublicKey))
	cli.Dag.Mux.Lock()
	copyDag := Copy(cli.Dag)
	cli.Dag.Mux.Unlock()
	copy(tx.LeftTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
	copy(tx.RightTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
	Crypto.PoW(&tx,2)
	buffer := serialize.SerializeData(tx)
	sign := GenerateSignature(buffer,privateKey)
	msg := GenerateMessage(buffer,sign)
	storage.AddTransaction(cli.Dag,tx,sign)
	BroadcastTransaction(msg,cli.Peers)
	return serial
}

func (cli *Client)handleConnection(conn net.Conn,url string) {
	for {
		buf := make([]byte,1024)
		l,err := conn.Read(buf)
		if(err != nil) {
			break;
		}
		cloudData := bytes.NewReader(cli.createTransaction(buf[:l]))
		http.Post(url,"application/json",cloudData)
	}
	defer conn.Close()
}


// RecieveSensorData is a method called in the main
func (cli *Client)RecieveSensorData(url string) {
	p := cli.Peers
	for {
		p.Mux.Lock()
		l := len(p.Fds)
		p.Mux.Unlock()
		if l > 0 {
			break
		}
		time.Sleep(2*time.Second)
	}
	fmt.Println("Starting Generating Transactions")
	listener,_ := net.Listen("tcp",":8050")
	for {
		conn,_ := listener.Accept()
		go cli.handleConnection(conn,url)
	}
}