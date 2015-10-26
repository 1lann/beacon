// Package protocol implements functions for reading, writing, decoding, and
// encoding data for the Minecraft protocol.
package protocol

import (
	"io"
)

type Stream struct {
	io.ReadWriter
}

type PacketStream struct {
	Stream
	reader *io.LimitedReader
}

type readWriter struct {
	io.Reader
	io.Writer
}

func (s PacketStream) ExhaustPacket() (int, error) {
	if s.reader.N == 0 {
		return 0, nil
	}

	return int(s.reader.N), s.ReadFull(make([]byte, s.reader.N))
}

func (s PacketStream) GetRemainingBytes() int {
	return int(s.reader.N)
}

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

func NewStream(readWriter io.ReadWriter) Stream {
	return Stream{readWriter}
}

// Returns decoded (little endian) data of length data.
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

func (s Stream) ReadFull(data []byte) error {
	_, err := io.ReadFull(s, data)
	if err != nil {
		return err
	}

	return nil
}

func (s Stream) WritePacket(p *Packet) error {
	lengthPacket := &Packet{}
	lengthPacket.WriteVarInt(len(p.Data))

	_, err := s.Write(append(lengthPacket.Data, p.Data...))
	if err != nil {
		return err
	}

	return nil
}
