package proto

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"math"
	"strings"

	"github.com/TyphoonMC/go.uuid"
)

type ConnReadWrite struct {
	rdr    io.Reader
	wtr    io.Writer
	buffer [16]byte
}

func (rdrwtr ConnReadWrite) ReadByte() (b byte, err error) {
	buff := rdrwtr.buffer[:1]
	if _, err = rdrwtr.rdr.Read(buff); err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (player *Player) ReadByte() (b byte, err error) {
	buff := player.io.buffer[:1]
	if _, err := io.ReadFull(player.conn, buff); err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (player *Player) ReadVarInt() (i int, err error) {
	v, err := binary.ReadUvarint(player.io)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func (player *Player) WriteVarInt(i int) (err error) {
	buff := player.io.buffer[:]
	length := binary.PutUvarint(buff, uint64(i))
	_, err = player.io.wtr.Write(buff[:length])
	return err
}

func (player *Player) ReadBool() (b bool, err error) {
	buff := player.io.buffer[:1]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return false, err
	}
	return buff[0] == 0x01, nil
}

func (player *Player) WriteBool(b bool) (err error) {
	buff := player.io.buffer[:1]
	if b {
		buff[0] = 0x01
	} else {
		buff[0] = 0x00
	}
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadUInt8() (i uint8, err error) {
	buff := player.io.buffer[:1]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (player *Player) WriteUInt8(i uint8) (err error) {
	buff := player.io.buffer[:1]
	buff[0] = i
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadUInt16() (i uint16, err error) {
	buff := player.io.buffer[:2]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff), nil
}

func (player *Player) WriteUInt16(i uint16) (err error) {
	buff := player.io.buffer[:2]
	binary.BigEndian.PutUint16(buff, i)
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadUInt32() (i uint32, err error) {
	buff := player.io.buffer[:4]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buff), nil
}

func (player *Player) WriteUInt32(i uint32) (err error) {
	buff := player.io.buffer[:4]
	binary.BigEndian.PutUint32(buff, i)
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadUInt64() (i uint64, err error) {
	buff := player.io.buffer[:8]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buff), nil
}

func (player *Player) WriteUInt64(i uint64) (err error) {
	buff := player.io.buffer[:8]
	binary.BigEndian.PutUint64(buff, i)
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadFloat32() (i float32, err error) {
	buff := player.io.buffer[:4]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(buff)), nil
}

func (player *Player) WriteFloat32(i float32) (err error) {
	buff := player.io.buffer[:4]
	binary.BigEndian.PutUint32(buff, math.Float32bits(i))
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadFloat64() (i float64, err error) {
	buff := player.io.buffer[:8]
	_, err = io.ReadFull(player.io.rdr, buff)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(buff)), nil
}

func (player *Player) WriteFloat64(i float64) (err error) {
	buff := player.io.buffer[:8]
	binary.BigEndian.PutUint64(buff, math.Float64bits(i))
	_, err = player.io.wtr.Write(buff)
	if err != nil {
		return err
	}
	return
}

func (player *Player) ReadString() (s string, err error) {
	length, err := player.ReadVarInt()
	if err != nil {
		return "", err
	}
	buffer := make([]byte, length)
	_, err = io.ReadFull(player.io.rdr, buffer)
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

func (player *Player) ReadStringLimited(max int) (s string, err error) {
	max = (max * 4) + 3

	length, err := player.ReadVarInt()
	if err != nil {
		return "", err
	}
	if length > max {
		log.Printf("invalid packet, too big")
		return "", nil
	}
	buffer := make([]byte, length)
	_, err = io.ReadFull(player.io.rdr, buffer)
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

func (player *Player) ReadNStringLimited(max int) (s string, read int, err error) {
	max = (max * 4) + 3

	length, err := player.ReadVarInt()
	buff := make([]byte, 8)
	read = binary.PutUvarint(buff, uint64(length))
	if err != nil {
		return "", read, err
	}
	if length > max {
		log.Printf("invalid packet, too big")
		return "", read, nil
	}
	buffer := make([]byte, length)
	_, err = io.ReadFull(player.io.rdr, buffer)
	if err != nil {
		return "", read + length, err
	}
	return string(buffer), read + length, nil
}

func (player *Player) WriteByteArray(data []byte) (err error) {
	_, err = player.io.wtr.Write(data)
	return err
}

func (player *Player) ReadByteArray(length int) (data []byte, err error) {
	data = make([]byte, length)
	_, err = player.io.rdr.Read(data)
	return data, err
}

func (player *Player) WriteString(s string) (err error) {
	buff := []byte(s)
	err = player.WriteVarInt(len(buff))
	if err != nil {
		return err
	}
	_, err = player.io.wtr.Write(buff)
	return err
}

func (player *Player) WriteStringRestricted(s string, max int) (err error) {
	buff := []byte(s)
	if len(buff) > max {
		buff = buff[:max]
	}
	err = player.WriteVarInt(len(buff))
	if err != nil {
		return err
	}
	_, err = player.io.wtr.Write(buff)
	return err
}

func (player *Player) WriteUUID(uid uuid.UUID) (err error) {
	_, err = player.io.wtr.Write(uid[:])
	return err
}

func JsonEscape(s string) string {
	str := strings.Replace(s, `\`, `\\`, -1)
	return strings.Replace(str, `"`, `\"`, -1)
}

func JsonMessage(s string) json.RawMessage {
	return json.RawMessage(`"` + JsonEscape(s) + `"`)
}
