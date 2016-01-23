package protocol

import (
	"encoding/binary"
	"errors"
)

var ErrInvalidData = errors.New("protocol: invalid data")

// ReadByte reads the next single byte from the stream.
func (s Stream) ReadByte() (byte, error) {
	data := make([]byte, 1)
	err := s.ReadFull(data)
	if err != nil {
		return 0, err
	}

	return data[0], nil
}

// ReadBoolean reads the next byte as a boolean from the stream.
func (s Stream) ReadBoolean() (bool, error) {
	b, err := s.ReadByte()
	if err != nil {
		return false, err
	}

	if b == 0x01 {
		return true, nil
	} else if b == 0x00 {
		return false, nil
	} else {
		return false, ErrInvalidData
	}
}

// ReadSignedByte reads the next byte as a signed byte (int8) from the stream.
func (s Stream) ReadSignedByte() (int8, error) {
	b, err := s.ReadByte()
	if err != nil {
		return 0, err
	}

	return int8(b), nil
}

// ReadInt16 reads the next 2 bytes as an int16 (a short) from the stream.
func (s Stream) ReadInt16() (int16, error) {
	b := make([]byte, 2)
	err := s.DecodeReadFull(b)
	if err != nil {
		return 0, err
	}

	return int16(binary.LittleEndian.Uint16(b)), nil
}

// ReadUInt16 reads the next 2 bytes as an uint16 (an unsigned short)
// from the stream.
func (s Stream) ReadUInt16() (uint16, error) {
	b := make([]byte, 2)
	err := s.DecodeReadFull(b)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(b), nil
}

// ReadInt32 reads the next 4 bytes as an int32 from the stream.
func (s Stream) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	err := s.DecodeReadFull(b)
	if err != nil {
		return 0, err
	}

	return int32(binary.LittleEndian.Uint32(b)), nil
}

// ReadInt32 reads the next 8 bytes as an int64 (a long) from the stream.
func (s Stream) ReadInt64() (int64, error) {
	b := make([]byte, 8)
	err := s.DecodeReadFull(b)
	if err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(b)), nil
}

// ReadFloat32 reads the next 4 bytes as a float32 (a single-precision float)
// from the stream.
func (s Stream) ReadFloat32() (float32, error) {
	var result float32
	err := binary.Read(s, binary.LittleEndian, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// ReadFloat32 reads the next 8 bytes as a float64 (a double-precision float)
// from the stream.
func (s Stream) ReadFloat64() (float64, error) {
	var result float64
	err := binary.Read(s, binary.LittleEndian, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// ReadString reads the next ReadVarInt bytes + the length of the VarInt
// and returns the string data from the stream.
func (s Stream) ReadString() (string, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", ErrInvalidData
	}

	data := make([]byte, length)
	err = s.ReadFull(data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ReadString reads the next couple of bytes which corresponds to a VarInt
// as an int from the stream.
func (s Stream) ReadVarInt() (int, error) {
	num, err := s.ReadVarInt64()
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

// ReadString reads the next couple of bytes which corresponds to a VarInt
// as an int64 from the stream.
//
// This code is taken from thinkofdeath's steven
// (github.com/thinkofdeath/steven).
func (s Stream) ReadVarInt64() (int64, error) {
	var size uint
	var num uint64

	for {
		b, err := s.ReadByte()
		if err != nil {
			return 0, err
		}

		num |= (uint64(b) & uint64(0x7F)) << (size * 7)
		size++
		if size > 10 {
			return 0, ErrInvalidData
		}

		if (b & 0x80) == 0 {
			break
		}
	}

	return int64(num), nil
}
