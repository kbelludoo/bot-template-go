package protocol

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Tipos de pacotes Engine.IO v4 / Socket.IO v4
const (
	EIOOpen    = "0"
	EIOClose   = "1"
	EIOPing    = "2"
	EIOPong    = "3"
	EIOMessage = "4"

	SIOConnect      = "40"
	SIODisconnect   = "41"
	SIOEvent        = "42"
	SIOAck          = "43"
	SIOConnectError = "44"
)

// SocketIOPacket representa um quadro decodificado do protocolo Engine.IO / Socket.IO v4
type SocketIOPacket struct {
	RawType string          // "0", "2", "3", "40", "42", "43", "44"
	AckID   int             // ID de correlação de ACK (-1 se ausente)
	Event   string          // Nome do evento em pacotes "42"
	Payload json.RawMessage // Dados do evento ou resposta de ACK
}

// ParseSocketIOPacket decodifica um frame de texto Engine.IO / Socket.IO v4 (padrão PokeLegacy)
func ParseSocketIOPacket(raw string) (SocketIOPacket, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SocketIOPacket{}, fmt.Errorf("pacote vazio")
	}

	pkt := SocketIOPacket{AckID: -1}

	switch {
	case raw == EIOPing || raw == EIOPong || raw == EIOClose:
		pkt.RawType = raw
		return pkt, nil

	case strings.HasPrefix(raw, EIOOpen):
		pkt.RawType = EIOOpen
		pkt.Payload = json.RawMessage(raw[1:])
		return pkt, nil

	case strings.HasPrefix(raw, SIOConnect):
		pkt.RawType = SIOConnect
		rest := strings.TrimPrefix(raw, SIOConnect)
		if idx := strings.IndexByte(rest, '{'); idx >= 0 {
			pkt.Payload = json.RawMessage(rest[idx:])
		}
		return pkt, nil

	case strings.HasPrefix(raw, SIOEvent):
		pkt.RawType = SIOEvent
		rest := strings.TrimPrefix(raw, SIOEvent)
		ackID, jsonPart := splitAckPrefix(rest)
		pkt.AckID = ackID

		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(jsonPart), &arr); err != nil || len(arr) == 0 {
			return pkt, fmt.Errorf("payload 42 inválido: %s", raw)
		}
		_ = json.Unmarshal(arr[0], &pkt.Event)
		if len(arr) > 1 {
			pkt.Payload = arr[1]
		}
		return pkt, nil

	case strings.HasPrefix(raw, SIOAck):
		pkt.RawType = SIOAck
		rest := strings.TrimPrefix(raw, SIOAck)
		ackID, jsonPart := splitAckPrefix(rest)
		pkt.AckID = ackID
		pkt.Payload = json.RawMessage(jsonPart)
		return pkt, nil

	default:
		pkt.RawType = raw[:1]
		return pkt, nil
	}
}

func splitAckPrefix(s string) (int, string) {
	idx := 0
	for idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
		idx++
	}
	if idx == 0 {
		return -1, s
	}
	id, err := strconv.Atoi(s[:idx])
	if err != nil {
		return -1, s
	}
	return id, s[idx:]
}

// FormatSocketIOConnect monta o pacote de handshake "40{...}"
func FormatSocketIOConnect(authPayload any) (string, error) {
	if authPayload == nil {
		return SIOConnect, nil
	}
	b, err := json.Marshal(authPayload)
	if err != nil {
		return "", err
	}
	return SIOConnect + string(b), nil
}

// FormatSocketIOEvent monta o pacote "42<ackID>[\"event\", payload]"
func FormatSocketIOEvent(event string, payload any, ackID int) (string, error) {
	arr := []any{event}
	if payload != nil {
		arr = append(arr, payload)
	}
	b, err := json.Marshal(arr)
	if err != nil {
		return "", err
	}
	if ackID >= 0 {
		return fmt.Sprintf("%s%d%s", SIOEvent, ackID, string(b)), nil
	}
	return SIOEvent + string(b), nil
}

var sha1ReleaseRegex = regexp.MustCompile(`(?i)clientRelease\s*[:=]\s*["']([a-f0-9]{40})["']`)

// DiscoverClientRelease extrai o hash SHA-1 de 40 caracteres de um bundle JS (padrão PokeLegacy)
func DiscoverClientRelease(bundleJS string) string {
	m := sha1ReleaseRegex.FindStringSubmatch(bundleJS)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
