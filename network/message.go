package network

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
)

// Protocol commands
const (
	CommandLength = 12

	CmdVersion   = "version"
	CmdGetBlocks = "getblocks"
	CmdInv       = "inv"
	CmdGetData   = "getdata"
	CmdBlock     = "block"
	CmdTx        = "tx"
	CmdPing      = "ping"
	CmdPong      = "pong"
)

// Message represents a network message
type Message struct {
	Command string
	Data    []byte
}

// VersionData represents version message payload
type VersionData struct {
	Version    int32
	BestHeight int32
	AddrFrom   string
}

// GetBlocksData represents getblocks message payload
type GetBlocksData struct {
	AddrFrom string
}

// InvData represents inventory message payload
type InvData struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

// GetDataData represents getdata message payload
type GetDataData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

// BlockData represents block message payload
type BlockData struct {
	AddrFrom string
	Block    []byte
}

// TxData represents transaction message payload
type TxData struct {
	AddrFrom    string
	Transaction []byte
}

// PingData represents ping message payload
type PingData struct {
	AddrFrom string
}

// PongData represents pong message payload
type PongData struct {
	AddrFrom string
}

// SerializeMessage serializes a message for network transmission
func SerializeMessage(msg Message) []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(msg)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeMessage deserializes a message from network data
func DeserializeMessage(data []byte) Message {
	var msg Message

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&msg)
	if err != nil {
		log.Panic(err)
	}

	return msg
}

// ReadMessage reads a complete message from a connection
func ReadMessage(r io.Reader) (Message, error) {
	// Read the message length first (4 bytes)
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBytes)
	if err != nil {
		return Message{}, err
	}

	// Convert bytes to length
	length := BytesToInt(lengthBytes)

	// Read the actual message
	msgBytes := make([]byte, length)
	_, err = io.ReadFull(r, msgBytes)
	if err != nil {
		return Message{}, err
	}

	return DeserializeMessage(msgBytes), nil
}

// WriteMessage writes a message to a connection
func WriteMessage(w io.Writer, msg Message) error {
	data := SerializeMessage(msg)

	// Write length first
	lengthBytes := IntToBytes(len(data))
	_, err := w.Write(lengthBytes)
	if err != nil {
		return err
	}

	// Write the actual message
	_, err = w.Write(data)
	return err
}

// IntToBytes converts an integer to byte slice (big-endian)
func IntToBytes(n int) []byte {
	result := make([]byte, 4)
	result[0] = byte(n >> 24)
	result[1] = byte(n >> 16)
	result[2] = byte(n >> 8)
	result[3] = byte(n)
	return result
}

// BytesToInt converts byte slice to integer (big-endian)
func BytesToInt(b []byte) int {
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

// GobEncode encodes data using gob
func GobEncode(data interface{}) []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// GobDecode decodes data using gob
func GobDecode(data []byte, v interface{}) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(v)
	if err != nil {
		log.Panic(err)
	}
}
