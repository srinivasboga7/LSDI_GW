package p2p

import "net"

import "bytes"

import "encoding/binary"

// Msg is the structure of all the msgs in the p2p network
type Msg struct {
	ID         uint32
	LenPayload uint32
	Payload    []byte
}

// ReadMsg reads the msg from the socket
func ReadMsg(conn net.Conn) (Msg, error) {
	bufHeader := make([]byte, 8)
	var msg Msg
	_, err := conn.Read(bufHeader)
	if err != nil {
		return msg, err
	}
	binary.Read(bytes.NewReader(bufHeader[:4]), binary.LittleEndian, msg.ID)
	binary.Read(bytes.NewReader(bufHeader[4:]), binary.LittleEndian, msg.LenPayload)
	bufPayload := make([]byte, msg.LenPayload)
	_, err = conn.Read(bufPayload)
	if err != nil {
		return msg, err
	}
	msg.Payload = bufPayload
	return msg, nil
}

// SendMsg sends the msg to the socket
func SendMsg(conn net.Conn, msg Msg, errc chan error) {
	// serialize the msg to be sent over the socket
	var b []byte
	// add the serialization
	_, err := conn.Write(b)
	if err != nil {
		errc <- err
	}
	return
}
