package protocol

import (
	"bytes"
	"testing"
)

func TestProtoReader_VarintAndString(t *testing.T) {
	// Mensagem simulada:
	// Field 1 (Varint): 150 (Level) -> Tag: (1 << 3) | 0 = 0x08, Val: 150 = 0x96, 0x01
	// Field 2 (LengthDelimited): "Hero" -> Tag: (2 << 3) | 2 = 0x12, Len: 4, Bytes: "Hero"
	// Field 3 (Fixed32): 100000 -> Tag: (3 << 3) | 5 = 0x1D, Bytes: 0xA0, 0x86, 0x01, 0x00
	buf := []byte{
		0x08, 0x96, 0x01,
		0x12, 0x04, 'H', 'e', 'r', 'o',
		0x1D, 0xA0, 0x86, 0x01, 0x00,
	}

	reader := NewProtoReader(buf)
	fields, err := reader.ReadAllFields()
	if err != nil {
		t.Fatalf("ReadAllFields failed: %v", err)
	}

	// Verificar Field 1 (Level = 150)
	f1, ok := fields[1]
	if !ok || len(f1) == 0 {
		t.Fatalf("Field 1 ausente")
	}
	if f1[0].Int64() != 150 {
		t.Errorf("Esperado Level 150, obteve %d", f1[0].Int64())
	}

	// Verificar Field 2 (Name = "Hero")
	f2, ok := fields[2]
	if !ok || len(f2) == 0 {
		t.Fatalf("Field 2 ausente")
	}
	if f2[0].String() != "Hero" {
		t.Errorf("Esperado Name 'Hero', obteve '%s'", f2[0].String())
	}

	// Verificar Field 3 (Gold = 100000)
	f3, ok := fields[3]
	if !ok || len(f3) == 0 {
		t.Fatalf("Field 3 ausente")
	}
	if f3[0].Fixed32 != 100000 {
		t.Errorf("Esperado Fixed32 100000, obteve %d", f3[0].Fixed32)
	}
}

func TestDecodeZigZag(t *testing.T) {
	// ZigZag encoding: 0 -> 0, -1 -> 1, 1 -> 2, -2 -> 3
	if DecodeZigZag(0) != 0 {
		t.Errorf("0 Zigzag falhou")
	}
	if DecodeZigZag(1) != -1 {
		t.Errorf("1 Zigzag esperado -1, obteve %d", DecodeZigZag(1))
	}
	if DecodeZigZag(2) != 1 {
		t.Errorf("2 Zigzag esperado 1, obteve %d", DecodeZigZag(2))
	}
	if DecodeZigZag(3) != -2 {
		t.Errorf("3 Zigzag esperado -2, obteve %d", DecodeZigZag(3))
	}
}

func TestDecodeVarint_EdgeCases(t *testing.T) {
	// Zero
	val, n, err := DecodeVarint([]byte{0x00})
	if err != nil || val != 0 || n != 1 {
		t.Errorf("Varint 0 falhou: val=%d n=%d err=%v", val, n, err)
	}

	// Multi-byte (300 = 0xAC, 0x02)
	val, n, err = DecodeVarint([]byte{0xAC, 0x02})
	if err != nil || val != 300 || n != 2 {
		t.Errorf("Varint 300 falhou: val=%d n=%d err=%v", val, n, err)
	}

	// Buffer muito curto
	_, _, err = DecodeVarint([]byte{0x80})
	if err != ErrBufferTooShort {
		t.Errorf("Esperava ErrBufferTooShort, obteve %v", err)
	}
	_ = bytes.Equal
}
