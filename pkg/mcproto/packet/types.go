package packet

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var (
	// ErrVarIntTooLong VarInt is too long to encode or is higher than a limit
	ErrVarIntTooLong = errors.New("VarInt is too long")
	// ErrStringLengthNegative StringLength cannot be negative
	ErrStringLengthNegative = errors.New("string length is below zero")
	// ErrStringTooLong String is too long to encode or is higher than a limit
	ErrStringTooLong = errors.New("string length is above maximum")
)

// ReadString reads an string from an io.Reader
func ReadString(r io.Reader) (val string, err error) {
	length, err := ReadVarInt(r)
	if err != nil {
		return
	}
	return ReadStringWithSize(r, length)
}

// ReadStringLimit reads an string from an io.Reader with a max length
func ReadStringLimit(r io.Reader, maxLength int) (val string, err error) {
	length, err := ReadVarInt(r)
	if err != nil {
		return
	}
	return ReadStringWithSizeLimit(r, length, maxLength)
}

// ReadStringWithSize reads an string with a defined length from an io.Reader
func ReadStringWithSize(r io.Reader, length int) (val string, err error) {
	return ReadStringWithSizeLimit(r, length, 1048576)
}

// ReadStringWithSizeLimit reads an string with a defined and max length from an io.Reader
func ReadStringWithSizeLimit(r io.Reader, length, maxLength int) (val string, err error) {
	if length < 0 {
		err = ErrStringLengthNegative
		return
	}
	if length > maxLength { // 2^(21-1)
		err = ErrStringTooLong
		return
	}
	bytes := make([]byte, length)
	_, err = r.Read(bytes)
	if err != nil {
		return
	}
	val = string(bytes)
	return
}

// WriteString writes an string to an io.Writer
func WriteString(w io.Writer, val string) (err error) {
	bytes := []byte(val)
	err = WriteVarInt(w, len(bytes))
	if err != nil {
		return
	}
	_, err = w.Write(bytes)
	return
}

// WriteString writes an string to an io.Writer
func WriteStringNew(w io.Writer, val string) (err error) {
	bytes := []byte(val)
	err = WriteVarIntNew(w, len(bytes))
	if err != nil {
		return
	}
	_, err = w.Write(bytes)
	return
}

// ReadVarInt reads a VarInt from an io.Reader
func ReadVarInt(r io.Reader) (result int, err error) {
	var bytes byte = 0
	var b byte
	for {
		b, err = ReadUint8(r)
		if err != nil {
			return
		}
		result |= int(uint(b&0x7F) << uint(bytes*7))
		bytes++
		if bytes > 5 {
			err = ErrVarIntTooLong
			return
		}
		if (b & 0x80) == 0x80 {
			continue
		}
		break
	}
	return
}

// WriteVarInt writes a VarInt to an io.Writer
func WriteVarInt(w io.Writer, val int) (err error) {
	for val >= 0x80 {
		err = WriteUint8(w, byte(val)|0x80)
		if err != nil {
			return
		}
		val >>= 7
	}
	err = WriteUint8(w, byte(val))
	return
}

// WriteVarIntNew writes a VarInt to an io.Writer
func WriteVarIntNew(w io.Writer, val int) (err error) {
	var buff [5]byte
	n := binary.PutUvarint(buff[:], uint64(val))
	_, err = w.Write(buff[:n])
	return
}

//
// func WriteUUID(writer io.Writer, val uuid.UUID) (err error) {
// 	_, err = writer.Write(val[:])
// 	return
// }
//
// func ReadUUID(reader io.Reader) (result uuid.UUID, err error) {
// 	bytes := make([]byte, 16)
// 	_, err = reader.Read(bytes)
// 	if err != nil {
// 		return
// 	}
// 	result, _ = uuid.FromBytes(bytes)
// 	return
// }

// ReadBool reads a boolean from an io.Reader
func ReadBool(r io.Reader) (val bool, err error) {
	uval, err := ReadUint8(r)
	if err != nil {
		return
	}
	val = uval != 0
	return
}

// WriteBool writes a boolean to an io.Writer
func WriteBool(w io.Writer, val bool) (err error) {
	if val {
		err = WriteUint8(w, 1)
	} else {
		err = WriteUint8(w, 0)
	}
	return
}

// ReadInt8 reads an int8 from an io.Reader
func ReadInt8(r io.Reader) (val int8, err error) {
	uval, err := ReadUint8(r)
	val = int8(uval)
	return
}

// WriteInt8 writes an int8 to an io.Writer
func WriteInt8(w io.Writer, val int8) (err error) {
	err = WriteUint8(w, uint8(val))
	return
}

// ReadUint8 reads an uint8 from an io.Reader
func ReadUint8(r io.Reader) (val uint8, err error) {
	var util [1]byte
	_, err = r.Read(util[:1])
	val = util[0]
	return
}

// WriteUint8 writes an uint8 to an io.Writer
func WriteUint8(w io.Writer, val uint8) (err error) {
	var util [1]byte
	util[0] = val
	_, err = w.Write(util[:1])
	return
}

// ReadInt16 reads an int16 from an io.Reader
func ReadInt16(r io.Reader) (val int16, err error) {
	uval, err := ReadUint16(r)
	val = int16(uval)
	return
}

// WriteInt16 writes an int16 to an io.Writer
func WriteInt16(w io.Writer, val int16) (err error) {
	err = WriteUint16(w, uint16(val))
	return
}

// ReadUint16 reads an uint16 from an io.Reader
func ReadUint16(r io.Reader) (val uint16, err error) {
	var util [2]byte
	_, err = r.Read(util[:2])
	val = binary.BigEndian.Uint16(util[:2])
	return
}

// WriteUint16 writes an uint16 to an io.Writer
func WriteUint16(w io.Writer, val uint16) (err error) {
	var util [2]byte
	binary.BigEndian.PutUint16(util[:2], val)
	_, err = w.Write(util[:2])
	return
}

// ReadInt32 reads an int32 from an io.Reader
func ReadInt32(r io.Reader) (val int32, err error) {
	uval, err := ReadUint32(r)
	val = int32(uval)
	return
}

// WriteInt32 writes an int32 to an io.Writer
func WriteInt32(w io.Writer, val int32) (err error) {
	err = WriteUint32(w, uint32(val))
	return
}

// ReadUint32 reads an uint32 from an io.Reader
func ReadUint32(r io.Reader) (val uint32, err error) {
	var util [4]byte
	_, err = r.Read(util[:4])
	val = binary.BigEndian.Uint32(util[:4])
	return
}

// WriteUint32 writes an uint32 to an io.Writer
func WriteUint32(w io.Writer, val uint32) (err error) {
	var util [4]byte
	binary.BigEndian.PutUint32(util[:4], val)
	_, err = w.Write(util[:4])
	return
}

// ReadInt64 reads an int64 from an io.Reader
func ReadInt64(r io.Reader) (val int64, err error) {
	uval, err := ReadUint64(r)
	val = int64(uval)
	return
}

// WriteInt64 writes an int64 to an io.Writer
func WriteInt64(w io.Writer, val int64) (err error) {
	err = WriteUint64(w, uint64(val))
	return
}

// ReadUint64 reads an uint64 from an io.Reader
func ReadUint64(r io.Reader) (val uint64, err error) {
	var util [8]byte
	_, err = r.Read(util[:8])
	val = binary.BigEndian.Uint64(util[:8])
	return
}

// WriteUint64 writes an uint64 to an io.Writer
func WriteUint64(w io.Writer, val uint64) (err error) {
	var util [8]byte
	binary.BigEndian.PutUint64(util[:8], val)
	_, err = w.Write(util[:8])
	return
}

// ReadFloat32 reads a float32 from an io.Reader
func ReadFloat32(r io.Reader) (val float32, err error) {
	ival, err := ReadUint32(r)
	val = math.Float32frombits(ival)
	return
}

// WriteFloat32 writes a float32 to an io.Writer
func WriteFloat32(w io.Writer, val float32) (err error) {
	return WriteUint32(w, math.Float32bits(val))
}

// ReadFloat64 reads a float32 from an io.Reader
func ReadFloat64(r io.Reader) (val float64, err error) {
	ival, err := ReadUint64(r)
	val = math.Float64frombits(ival)
	return
}

// WriteFloat64 writes a float64 to an io.Writer
func WriteFloat64(w io.Writer, val float64) (err error) {
	return WriteUint64(w, math.Float64bits(val))
}
