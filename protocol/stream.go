// Package protocol implements functions for reading, writing, decoding, and
// encoding data for the Minecraft protocol.
package protocol

import (
	"io"
)

// A Stream represents a two-way stream of bytes to and from the client.
type Stream struct {
	io.ReadWriter
}

// A PacketStream is a subset of a Stream which is limited to being only able
// to read a single packet from the stream.
type PacketStream struct {
	Stream
	reader *io.LimitedReader
}

type readWriter struct {
	io.Reader
	io.Writer
}

// ExhaustPacket reads all the remaining data from the PacketStream, so the
// cursor of the Stream is at the start of the next packet.
func (s PacketStream) ExhaustPacket() (int, error) {
	if s.reader.N == 0 {
		return 0, nil
	}

	bytesRemaining := int(s.reader.N)

	return bytesRemaining, s.ReadFull(make([]byte, bytesRemaining))
}

// GetRemainingBytes returns the number of remaining bytes in the PacketStream
// for the current packet.
func (s PacketStream) GetRemainingBytes() int {
	return int(s.reader.N)
}

// GetPacketStream reads the next VarInt, and creates a PacketStream limited
// by the VarInt representing the entirety of the packet.
func (s Stream) GetPacketStream() (PacketStream, int, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return PacketStream{}, 0, err
	}

	if length == 0 {
		return PacketStream{}, 0, ErrInvalidData
	}

	limitedReader := &io.LimitedReader{R: s, N: int64(length)}

	return PacketStream{Stream{readWriter{limitedReader, s}}, limitedReader},
		length, nil
}

// NewStream creates a new Stream from a io.ReadWriter such as from a net.Conn
func NewStream(readWriter io.ReadWriter) Stream {
	return Stream{readWriter}
}

// DecodeReadFull returns decoded (little endian) data of len(data), or what's
// left of the Stream if there is less data remaining available for the stream.
func (s Stream) DecodeReadFull(data []byte) error {
	_, err := io.ReadFull(s, data)
	if err != nil {
		return err
	}

	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return nil
}

// ReadFull returns raw data (big endian) of len(data), or what's left
// of the Stream if there is less data remaining available for the stream.
func (s Stream) ReadFull(data []byte) error {
	_, err := io.ReadFull(s, data)
	if err != nil {
		return err
	}

	return nil
}

// WritePacket writes the length of the Packet as a VarInt, and the Packet's
// Data (payload) to the stream.
func (s Stream) WritePacket(p *Packet) error {
	lengthPacket := &Packet{}
	lengthPacket.WriteVarInt(len(p.Data))

	_, err := s.Write(append(lengthPacket.Data, p.Data...))
	if err != nil {
		return err
	}

	return nil
}
