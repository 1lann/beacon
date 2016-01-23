package protocol

import (
	"encoding/binary"
)

// A Packet represents a packet in the Minecraft protocol that is to be sent
// to the client. It is not used for receiving or reading packets.
type Packet struct {
	Data []byte
}

// NewPacketWithId returns a new Packet with the Id written to the start
// of the Packet.
func NewPacketWithId(id int) *Packet {
	p := &Packet{}
	p.WriteVarInt(id)
	return p
}

// Write writes arbitrary bytes to the Packet.
func (p *Packet) Write(data []byte) (int, error) {
	p.Data = append(p.Data, data...)
	return len(data), nil
}

// WriteByte writes a single byte to the Packet.
func (p *Packet) WriteByte(data byte) error {
	p.Data = append(p.Data, data)
	return nil
}

// WriteBoolean writes a boolean to the Packet.
func (p *Packet) WriteBoolean(data bool) {
	if data {
		p.Data = append(p.Data, 0x01)
	} else {
		p.Data = append(p.Data, 0x00)
	}
}

// WriteSignedByte writes an int8 to the Packet.
func (p *Packet) WriteSignedByte(data int8) {
	p.Data = append(p.Data, byte(data))
}

// WriteInt16 writes an int16 (a short) to the Packet.
func (p *Packet) WriteInt16(data int16) {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, uint16(data))
	p.Data = append(p.Data, result...)
}

// WriteUInt16 writes an uint16 (an unsigned short) to the Packet.
func (p *Packet) WriteUInt16(data uint16) {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, data)
	p.Data = append(p.Data, result...)
}

// WriteInt32 writes an int32 to the Packet.
func (p *Packet) WriteInt32(data int32) {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(data))
	p.Data = append(p.Data, result...)
}

// WriteInt64 writes an int64 (a long) to the Packet.
func (p *Packet) WriteInt64(data int64) {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(data))
	p.Data = append(p.Data, result...)
}

// WriteFloat32 writes a float32 (a single-precision float) to the Packet.
func (p *Packet) WriteFloat32(data float32) {
	binary.Write(p, binary.BigEndian, data)
	return
}

// WriteFloat64 writes a float64 (a double-precision float) to the Packet.
func (p *Packet) WriteFloat64(data float64) {
	binary.Write(p, binary.BigEndian, data)
	return
}

// WriteString writes a VarInt which is the length of the string, and the
// string itself to the Packet.
func (p *Packet) WriteString(data string) {
	p.WriteVarInt(len(data))
	p.Data = append(p.Data, []byte(data)...)
}

// WriteVarInt writes an int as a VarInt to the Packet.
func (p *Packet) WriteVarInt(data int) {
	p.WriteVarInt64(int64(data))
}

// WriteVarInt64 writes an int64 as a VarInt to the Packet.
//
// This code is taken from thinkofdeath's steven
// (github.com/thinkofdeath/steven).
func (p *Packet) WriteVarInt64(data int64) {
	ui := uint64(data)
	for {
		if (ui & ^uint64(0x7F)) == 0 {
			p.Data = append(p.Data, byte(ui))
			return
		}
		p.Data = append(p.Data, byte((ui&0x7F)|0x80))
		ui >>= 7
	}
}
