package p2p

import "net"

import "bytes"

import "encoding/binary"

import "time"

const (
	rwDeadline = 2 * time.Second
)

// Msg is the structure of all the msgs in the p2p network
type Msg struct {
	ID         uint32
	LenPayload uint32
	Payload    []byte
}

func (msg *Msg) encode() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, msg.ID)
	binary.Write(buf, binary.LittleEndian, msg.LenPayload)
	szx := append(buf.Bytes(), msg.Payload...)
	return szx
}

// ReadMsg reads the msg from the socket
func ReadMsg(conn net.Conn) (Msg, error) {
	conn.SetReadDeadline(time.Now().Add(rwDeadline))
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
func SendMsg(conn net.Conn, msg Msg) error {
	// serialize the msg to be sent over the socket
	conn.SetWriteDeadline(time.Now().Add(rwDeadline)) // timeout
	_, err := conn.Write(msg.encode())
	return err
}
