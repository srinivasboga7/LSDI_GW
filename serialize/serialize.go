package serialize

import (
	"reflect"
	"bytes"
	"encoding/binary"
	dt "GO-DAG/DataTypes"
	"fmt"
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
	// encode based on type as binary package doesn't support strings
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

func Deserialize(b []byte) (dt.Transaction,[]byte) {
	// only a temporary method will change to include signature and other checks
	l := binary.LittleEndian.Uint32(b[:4])
	payload := b[4:l+4]
	signature := b[l+4:]
	r := bytes.NewReader(payload)
	var tx dt.Transaction
	err := binary.Read(r,binary.LittleEndian,&tx)
	if err != nil { 
		fmt.Println(err)
	}
	return tx,signature
}