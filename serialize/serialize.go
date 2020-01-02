package serialize

import (
	"reflect"
	"bytes"
	"math/big"
	"encoding/asn1"
	"encoding/binary"
	//"encoding/json"
	dt "GO-DAG/datatypes"
	"fmt"
)

//EncodeToHex converts byte slice to the string
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

//DecodeToBytes converts string to byte slice
func DecodeToBytes(data string) []byte {
	b,_ := hex.DecodeString(data)
	return b
}

//SerializeData serializes transaction to byte slice
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

//EncodeToBytes Converts any type to byte slice, supports strings,integers etc.
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

//DeserializeTransaction Converts back byte slice to transaction
func DeserializeTransaction(b []byte) (dt.Transaction,[]byte) {
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

//canonicalizeInt Converts bigint to byte slice
func canonicalizeInt(val *big.Int) []byte {
	b := val.Bytes()
	if len(b) == 0 {
		b = []byte{0x00}
	}
	if b[0]&0x80 != 0 {
		paddedBytes := make([]byte, len(b)+1)
		copy(paddedBytes[1:], b)
		b = paddedBytes
	}
	return b
}

// // Deserializing signature from DER encoded format

// func PointsFromDER(der []byte) (*big.Int,*big.Int) {

// 	R, S := &big.Int{}, &big.Int{}

// 	data := asn1.RawValue{}
// 	if _, err := asn1.Unmarshal(der, &data); err != nil {
// 		panic(err.Error())
// 	}

// 	// The format of our DER string is 0x02 + rlen + r + 0x02 + slen + s
// 	rLen := data.Bytes[1] // The entire length of R + offset of 2 for 0x02 and rlen
// 	r := data.Bytes[2 : rLen+2]
// 	// Ignore the next 0x02 and slen bytes and just take the start of S to the end of the byte array
// 	s := data.Bytes[rLen+4:]

// 	R.SetBytes(r)
// 	S.SetBytes(s)

// 	return R,S

// }

// // serializing signature to DER encoded format

// func PointsToDER(r,s *big.Int) []byte {

// 	// Ensure the encoded bytes for the r and s values are canonical and
// 	// thus suitable for DER encoding.
// 	rb := canonicalizeInt(r)
// 	sb := canonicalizeInt(s)

// 	// total length of returned signature is 1 byte for each magic and
// 	// length (6 total), plus lengths of r and s
// 	length := 6 + len(rb) + len(sb)
// 	b := make([]byte, length)

// 	b[0] = 0x30
// 	b[1] = byte(length - 2)
// 	b[2] = 0x02
// 	b[3] = byte(len(rb))
// 	offset := copy(b[4:], rb) + 4
// 	b[offset] = 0x02
// 	b[offset+1] = byte(len(sb))
// 	copy(b[offset+2:], sb)
// 	return b
// } 