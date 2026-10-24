package protocol

import (
	"bytes"
	"testing"
)

func TestProtoWriterAndFatigueState(t *testing.T) {
	w := NewProtoWriter()
	w.WriteVarintField(1, 820)    // fatigue_milli = 82.0%
	w.WriteVarintField(2, 100000) // window_start_ms
	w.WriteVarintField(3, 50000)  // window_farm_ms
	w.WriteVarintField(4, 90000)  // farm_available_ms
	w.WriteFloat64Field(5, 34.5)  // online_kpm

	pkt, err := ParseFatigueState(w.Bytes())
	if err != nil {
		t.Fatalf("erro ao decodificar FatigueState: %v", err)
	}
	if pkt.FatigueMilli != 820 || pkt.FarmAvailableMs != 90000 || pkt.OnlineKPM != 34.5 {
		t.Fatalf("FatiguePacket inesperado: %+v", pkt)
	}
}

func TestSocketIOPacketAndRelease(t *testing.T) {
	evStr, err := FormatSocketIOEvent("battle:start", map[string]int{"zoneId": 7}, 12)
	if err != nil {
		t.Fatalf("erro ao formatar 42: %v", err)
	}
	parsed, err := ParseSocketIOPacket(evStr)
	if err != nil {
		t.Fatalf("erro ao parsear 42: %v", err)
	}
	if parsed.RawType != SIOEvent || parsed.AckID != 12 || parsed.Event != "battle:start" {
		t.Fatalf("pacote Socket.IO inesperado: %+v", parsed)
	}

	js := `const cfg = { clientRelease: "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4" };`
	if rel := DiscoverClientRelease(js); rel != "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4" {
		t.Fatalf("release não extraído: %s", rel)
	}
}

func TestColyseusAndTRPC(t *testing.T) {
	joinFrame := []byte{ColyseusOpJoinRoom, 5, 't', 'o', 'k', 'e', 'n', 6, 's', 'c', 'h', 'e', 'm', 'a'}
	tok, ser, ack, err := ParseColyseusJoinRoom(joinFrame)
	if err != nil || tok != "token" || ser != "schema" || !bytes.Equal(ack, []byte{10}) {
		t.Fatalf("falha em ParseColyseusJoinRoom: tok=%s ser=%s err=%v", tok, ser, err)
	}

	roomData := EncodeColyseusRoomData("combat", []byte{0x80})
	if roomData[0] != ColyseusOpRoomData || roomData[1] != (0xa0|6) {
		t.Fatalf("EncodeColyseusRoomData inválido: %v", roomData)
	}

	rawResp := []byte(`[{"result":{"data":{"json":{"gold":999}}}}]`)
	var out struct {
		Gold int `json:"gold"`
	}
	if err := UnwrapTRPCBatchResponse(rawResp, &out); err != nil || out.Gold != 999 {
		t.Fatalf("falha em UnwrapTRPCBatchResponse: %v (%+v)", err, out)
	}
}
