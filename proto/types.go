package proto

import (
	"encoding/json"
)

const defaultRPCVersion = "1"

const (
	handshakeOpcode uint32 = iota
	frameOpcode
	closeOpcode
	pingOpcode
	pongOpcode
)

const (
	eventTypeReady = "READY"
	eventTypeError = "ERROR"
)

type handshakePacket struct {
	Version  string `json:"v"`
	ClientID string `json:"client_id"`
}

type framePacket struct {
	Command string          `json:"cmd"`
	Data    json.RawMessage `json:"data"`
	Args    json.RawMessage `json:"args"`
	Event   string          `json:"evt"`
	Nonce   string          `json:"nonce"`
}
