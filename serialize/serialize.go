package serialize

import (
	dt "GO-DAG/DataTypes"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
)

//EncodeToHex converts byte slice to the string
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

//DecodeToBytes converts string to byte slice
func DecodeToBytes(data string) []byte {
	b, _ := hex.DecodeString(data)
	return b
}

//Encode serializes transaction to byte slice
func Encode(t interface{}) []byte {
	// iterating over a struct is painful in golang
	var b []byte
	v := reflect.ValueOf(&t).Elem()
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		b = append(b, EncodeToBytes(value.Interface())...)
	}
	return b
}

//EncodeToBytes Converts any type to byte slice, supports strings,integers etc.
func EncodeToBytes(x interface{}) []byte {
	// encode based on type as binary package doesn't support strings
	switch x.(type) {
	case string:
		str := x.(string)
		return []byte(str)
	default:
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, x)
		return buf.Bytes()
	}
}

//DeserializeTransaction Converts back byte slice to transaction
func DeserializeTransaction(payload []byte, lenPayload uint32, Type string) (interface{}, []byte) {
	signature := payload[lenPayload-72:]
	r := bytes.NewReader(payload[:lenPayload-72])
	switch Type {
	case "DataTx":
		var tx dt.Transaction
		err := binary.Read(r, binary.LittleEndian, &tx)
		if err != nil {
			fmt.Println(err)
		}
		return tx, signature
	case "ShardSignal":
		var tx dt.ShardSignal
		err := binary.Read(r, binary.LittleEndian, &tx)
		if err != nil {
			fmt.Println(err)
		}
		return tx, signature
	case "ShardTx":
		var tx dt.ShardTransaction
		err := binary.Read(r, binary.LittleEndian, &tx)
		if err != nil {
			fmt.Println(err)
		}
		return tx, signature
	default:
		return nil, nil
	}
}
