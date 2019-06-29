package client

import (
	"math/rand"
	dt "GO-DAG/DataTypes"
	"time"
	"encoding/binary"
	"bytes"
)

func SerializeData(t dt.Transaction) []byte {
	// iterating over a struct is painful in golang
	var b []byte
	v := reflect.ValueOf(&t).Elem()
	for i := 0; i < v.NumField() ;i++ {
		value := v.Field(i)
		b = append(b,EncodeToBytes(value.Interface())...)
	}
	return b
}

func EncodeToBytes(x interface{}) []byte {
	switch t := x.(type) {
	case string :
		return []byte(x)
	default :
		buf := new(bytes.Buffer)
		err := binary.Write(buf,binary.LittleEndian, v)
		return buf.Bytes()
	}
}


func BroadcastTransaction(b []byte, p dt.Peers) {
	p.Mux.Lock()
	buf := b
	for _,conn := range p.Fds {
		conn.Write(buf)
	}
	p.Mux.Unlock()
}


func SimulateClient(p dt.Peers) {

	var tx dt.Transaction

	for {
		tx.Timestamp = time.Now().Unix()
		tx.Value = rand.Float64()
		buffer := SerializeData(tx)
		BroadcastTransaction(buffer,p)
		time.Sleep(2*time.Second)
	}
}