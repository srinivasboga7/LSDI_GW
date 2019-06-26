package main

/*
import(
	"GO-DAG/server"
	"GO-DAG/client"
	dt "GO-DAG/DataTypes"
	"net"
	"time"
)


func main() {
	go server.StartServer()
	time.Sleep(time.Second)
	var p dt.Peers
	p.Fds = make([]net.Conn,1)
	conn,_ := net.Dial("tcp","127.0.0.1:9000")
	p.Fds[0] = conn
	client.SimulateClient(p)
}
*/

import(
	"fmt"
	"bytes"
	"encoding/binary"
	"reflect"
	//"encoding/hex"
)

type data struct {
	Timestamp float64
	Value float64
	Nonce uint32
	From [8]byte
}

func SerializeData(t data) []byte {
	// too much overhead has to change
	/*
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	e.Encode(t)
	buf := b.Bytes()
	*/
	var b []byte
	v := reflect.ValueOf(&t).Elem()
	for i := 0; i < v.NumField() ;i++ {
		value := v.Field(i)
		b = append(b,EncodeToBytes(value.Interface())...)
	}
	return b
}

func EncodeToBytes(x interface{}) []byte {
	switch x.(type) {
	case string :
		str := x.(string)
		return []byte(str)
	default :
		buf := new(bytes.Buffer)
		binary.Write(buf,binary.LittleEndian, x)
		return buf.Bytes()
	}
}

func Deserialize(b []byte) data {
	r := bytes.NewReader(b)
	var tx data
	err := binary.Read(r,binary.LittleEndian,&tx)
	fmt.Println(err)
	return tx
}

func main() {
	var d data
	d.Timestamp = 34235.54
	d.Value = 1352346.4
	d.Nonce = 9083
	byt := []byte("srinivas")
	copy(d.From[:],byt)
	b := SerializeData(d)
	fmt.Println(len(b))
	tx := Deserialize(b)
	fmt.Println(tx)
}