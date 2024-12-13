package model

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/gavv/monotime"
)

type RawByte byte

type PacketRecord struct {
	Stream []byte
	Time   time.Time
}

// NewPacketRecord contains packet bytes
func NewPacketRecord(
	stream []byte,
	len uint32,
	ts time.Time,
) *PacketRecord {
	pr := PacketRecord{}
	pr.Time = ts
	pr.Stream = make([]byte, len)
	pr.Stream = stream
	return &pr
}

// ReadRawPacket reads a PacketRecord from a binary source, in LittleEndian order
func ReadRawPacket(reader io.Reader) (*PacketRecord, error) {
	var pr PacketRecord
	currentTime := time.Now()
	monotonicTimeNow := monotime.Now()
	getLen := make([]byte, 4)
	packetTimestamp := make([]byte, 8)
	// Read IfIndex and discard it: To be used in other usecases
	_ = binary.Read(reader, binary.LittleEndian, make([]byte, 4))
	// Read Length of packet
	_ = binary.Read(reader, binary.LittleEndian, getLen)
	pr.Stream = make([]byte, binary.LittleEndian.Uint32(getLen))
	// Read TimeStamp of packet
	_ = binary.Read(reader, binary.LittleEndian, packetTimestamp)
	// The assumption is monotonic time should be as close to time recorded by ebpf.
	// The difference is considered the delta time from current time.
	tsDelta := time.Duration(uint64(monotonicTimeNow) - binary.LittleEndian.Uint64(packetTimestamp))
	pr.Time = currentTime.Add(-tsDelta)

	err := binary.Read(reader, binary.LittleEndian, &pr.Stream)
	return &pr, err
}
