package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// Tipos de fio (Wire Types) do Protobuf
const (
	WireVarint          = 0
	WireFixed64         = 1
	WireLengthDelimited = 2
	WireStartGroup      = 3 // Deprecated
	WireEndGroup        = 4 // Deprecated
	WireFixed32         = 5
)

var (
	ErrBufferTooShort = errors.New("buffer curto demais")
	ErrOverflow       = errors.New("varint overflow (> 64 bits)")
)

// DecodeVarint decodifica um varint uint64 a partir de uma fatia de bytes
func DecodeVarint(buf []byte) (uint64, int, error) {
	var val uint64
	var shift uint
	for i, b := range buf {
		if i >= 10 {
			return 0, 0, ErrOverflow
		}
		val |= uint64(b&0x7F) << shift
		if (b & 0x80) == 0 {
			return val, i + 1, nil
		}
		shift += 7
	}
	return 0, 0, ErrBufferTooShort
}

// DecodeZigZag decodifica inteiros com sinal codificados com Sint32 ou Sint64
func DecodeZigZag(n uint64) int64 {
	return int64((n >> 1) ^ -(n & 1))
}

// DecodeTag decodifica o número do campo e o wire type de um cabeçalho Protobuf
func DecodeTag(buf []byte) (fieldNumber int, wireType int, n int, error error) {
	v, n, err := DecodeVarint(buf)
	if err != nil {
		return 0, 0, 0, err
	}
	wireType = int(v & 0x07)
	fieldNumber = int(v >> 3)
	return fieldNumber, wireType, n, nil
}

// ProtoField representa um campo decodificado de mensagem binária Protobuf
type ProtoField struct {
	FieldNumber int
	WireType    int
	VarintValue uint64
	Fixed32     uint32
	Fixed64     uint64
	Bytes       []byte
}

// Int64 devolve o valor como int64
func (f ProtoField) Int64() int64 {
	return int64(f.VarintValue)
}

// String devolve o valor como string UTF-8
func (f ProtoField) String() string {
	return string(f.Bytes)
}

// ProtoReader leitor leve sem alocação para mensagens Protobuf arbitrárias
type ProtoReader struct {
	buf []byte
	pos int
}

// NewProtoReader inicializa o leitor
func NewProtoReader(buf []byte) *ProtoReader {
	return &ProtoReader{buf: buf, pos: 0}
}

// HasNext verifica se ainda há bytes para ler
func (r *ProtoReader) HasNext() bool {
	return r.pos < len(r.buf)
}

// NextField lê o próximo campo do buffer Protobuf
func (r *ProtoReader) NextField() (ProtoField, error) {
	if r.pos >= len(r.buf) {
		return ProtoField{}, io.EOF
	}

	fn, wt, n, err := DecodeTag(r.buf[r.pos:])
	if err != nil {
		return ProtoField{}, err
	}
	r.pos += n

	field := ProtoField{
		FieldNumber: fn,
		WireType:    wt,
	}

	switch wt {
	case WireVarint:
		v, vn, err := DecodeVarint(r.buf[r.pos:])
		if err != nil {
			return field, err
		}
		r.pos += vn
		field.VarintValue = v

	case WireFixed32:
		if r.pos+4 > len(r.buf) {
			return field, ErrBufferTooShort
		}
		field.Fixed32 = binary.LittleEndian.Uint32(r.buf[r.pos : r.pos+4])
		r.pos += 4

	case WireFixed64:
		if r.pos+8 > len(r.buf) {
			return field, ErrBufferTooShort
		}
		field.Fixed64 = binary.LittleEndian.Uint64(r.buf[r.pos : r.pos+8])
		r.pos += 8

	case WireLengthDelimited:
		lenBytes, ln, err := DecodeVarint(r.buf[r.pos:])
		if err != nil {
			return field, err
		}
		r.pos += ln
		end := r.pos + int(lenBytes)
		if end > len(r.buf) {
			return field, ErrBufferTooShort
		}
		field.Bytes = r.buf[r.pos:end]
		r.pos = end

	default:
		return field, fmt.Errorf("wire type não suportado: %d", wt)
	}

	return field, nil
}

// ReadAllFields extrai todos os campos em um mapa indexado por fieldNumber
func (r *ProtoReader) ReadAllFields() (map[int][]ProtoField, error) {
	result := make(map[int][]ProtoField)
	for r.HasNext() {
		f, err := r.NextField()
		if err != nil {
			return result, err
		}
		result[f.FieldNumber] = append(result[f.FieldNumber], f)
	}
	return result, nil
}

// EncodeVarint codifica um uint64 em formato Varint do Protobuf
func EncodeVarint(v uint64) []byte {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], v)
	out := make([]byte, n)
	copy(out, buf[:n])
	return out
}

// ProtoWriter construtor leve sem schema (.proto) para mensagens Protobuf
type ProtoWriter struct {
	buf []byte
}

// NewProtoWriter cria um novo escritor Protobuf
func NewProtoWriter() *ProtoWriter {
	return &ProtoWriter{buf: make([]byte, 0, 64)}
}

func (w *ProtoWriter) writeTag(fieldNumber int, wireType int) {
	tag := uint64((fieldNumber << 3) | (wireType & 0x07))
	w.buf = append(w.buf, EncodeVarint(tag)...)
}

// WriteVarintField escreve um campo Varint (int32, int64, uint64, bool, enum)
func (w *ProtoWriter) WriteVarintField(fieldNumber int, val uint64) {
	w.writeTag(fieldNumber, WireVarint)
	w.buf = append(w.buf, EncodeVarint(val)...)
}

// WriteBytesField escreve um campo length-delimited ([]byte ou submensagem)
func (w *ProtoWriter) WriteBytesField(fieldNumber int, data []byte) {
	w.writeTag(fieldNumber, WireLengthDelimited)
	w.buf = append(w.buf, EncodeVarint(uint64(len(data)))...)
	w.buf = append(w.buf, data...)
}

// WriteStringField escreve uma string UTF-8
func (w *ProtoWriter) WriteStringField(fieldNumber int, s string) {
	w.WriteBytesField(fieldNumber, []byte(s))
}

// WriteSubMessage escreve uma submensagem construída com outro ProtoWriter
func (w *ProtoWriter) WriteSubMessage(fieldNumber int, sub *ProtoWriter) {
	if sub == nil {
		return
	}
	w.WriteBytesField(fieldNumber, sub.Bytes())
}

// WriteFixed32Field escreve 4 bytes little-endian
func (w *ProtoWriter) WriteFixed32Field(fieldNumber int, val uint32) {
	w.writeTag(fieldNumber, WireFixed32)
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], val)
	w.buf = append(w.buf, b[:]...)
}

// WriteFloat32Field escreve um float32 IEEE-754
func (w *ProtoWriter) WriteFloat32Field(fieldNumber int, val float32) {
	w.WriteFixed32Field(fieldNumber, math.Float32bits(val))
}

// WriteFixed64Field escreve 8 bytes little-endian
func (w *ProtoWriter) WriteFixed64Field(fieldNumber int, val uint64) {
	w.writeTag(fieldNumber, WireFixed64)
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], val)
	w.buf = append(w.buf, b[:]...)
}

// WriteFloat64Field escreve um float64 IEEE-754
func (w *ProtoWriter) WriteFloat64Field(fieldNumber int, val float64) {
	w.WriteFixed64Field(fieldNumber, math.Float64bits(val))
}

// Bytes retorna o buffer binário serializado
func (w *ProtoWriter) Bytes() []byte {
	return w.buf
}

// FatiguePacket representa o estado de fadiga decodificado de jogos Protobuf (ex: Leveling Idle)
type FatiguePacket struct {
	FatigueMilli    int64   // 0..1000 (ex: 800 = 80.0% de fadiga)
	WindowStartMs   int64   // Timestamp de início da janela
	WindowFarmMs    int64   // Tempo farmado na janela atual
	FarmAvailableMs int64   // Tempo restante disponível para farm (ms)
	OnlineKPM       float64 // Abates por minuto medidos pelo servidor
}

// ParseFatigueState decodifica um sub-pacote Protobuf FatigueState (campos 1..5)
func ParseFatigueState(buf []byte) (FatiguePacket, error) {
	r := NewProtoReader(buf)
	var pkt FatiguePacket
	for r.HasNext() {
		f, err := r.NextField()
		if err != nil {
			return pkt, err
		}
		switch f.FieldNumber {
		case 1:
			pkt.FatigueMilli = int64(f.VarintValue)
		case 2:
			pkt.WindowStartMs = int64(f.VarintValue)
		case 3:
			pkt.WindowFarmMs = int64(f.VarintValue)
		case 4:
			pkt.FarmAvailableMs = int64(f.VarintValue)
		case 5:
			if f.WireType == WireFixed64 {
				pkt.OnlineKPM = math.Float64frombits(f.Fixed64)
			} else if f.WireType == WireFixed32 {
				pkt.OnlineKPM = float64(math.Float32frombits(f.Fixed32))
			}
		}
	}
	return pkt, nil
}
