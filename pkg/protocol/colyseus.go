package protocol

import (
	"fmt"
)

// Opcodes binários oficiais do Colyseus v0.16 (padrão Baiak Idle)
const (
	ColyseusOpHandshake      byte = 9
	ColyseusOpJoinRoom       byte = 10
	ColyseusOpError          byte = 11
	ColyseusOpLeaveRoom      byte = 12
	ColyseusOpRoomData       byte = 13
	ColyseusOpRoomState      byte = 14
	ColyseusOpRoomStatePatch byte = 15
	ColyseusOpPing           byte = 18
	ColyseusOpPong           byte = 19
)

// ParseColyseusJoinRoom decodifica o quadro binário OpJoinRoom (10) e retorna o reconnectionToken, serializerID e o pacote de ACK obrigatório ([]byte{10})
func ParseColyseusJoinRoom(frame []byte) (reconnectionToken string, serializerID string, ackFrame []byte, err error) {
	if len(frame) < 2 || frame[0] != ColyseusOpJoinRoom {
		return "", "", nil, fmt.Errorf("frame inválido para OpJoinRoom")
	}
	offset := 1
	tokenLen := int(frame[offset])
	offset++
	if offset+tokenLen > len(frame) {
		return "", "", nil, ErrBufferTooShort
	}
	reconnectionToken = string(frame[offset : offset+tokenLen])
	offset += tokenLen

	if offset < len(frame) {
		serLen := int(frame[offset])
		offset++
		if offset+serLen <= len(frame) {
			serializerID = string(frame[offset : offset+serLen])
		}
	}
	return reconnectionToken, serializerID, []byte{ColyseusOpJoinRoom}, nil
}

// EncodeColyseusRoomData monta o quadro binário OpRoomData (13) com cabeçalho msgpack fixstr/str8 e payload msgpack
func EncodeColyseusRoomData(msgType string, msgpackPayload []byte) []byte {
	typeBytes := []byte(msgType)
	var header []byte
	if len(typeBytes) <= 31 {
		// msgpack fixstr (0xa0 | len)
		header = []byte{ColyseusOpRoomData, byte(0xa0 | len(typeBytes))}
	} else {
		// msgpack str8 (0xd9, len)
		header = []byte{ColyseusOpRoomData, 0xd9, byte(len(typeBytes))}
	}
	out := make([]byte, 0, len(header)+len(typeBytes)+len(msgpackPayload))
	out = append(out, header...)
	out = append(out, typeBytes...)
	out = append(out, msgpackPayload...)
	return out
}
