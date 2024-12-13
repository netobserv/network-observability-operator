package packets

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// PCAP Magic number is fixed for each endianness.
const pcapMagicNumber = 0xA1B2C3D4
const versionMajor = 2
const versionMinor = 4
const nanosPerMicro = 1000

func GetPCAPFileHeader(snaplen uint32, linktype layers.LinkType) []byte {
	var buf [24]byte
	binary.LittleEndian.PutUint32(buf[0:4], pcapMagicNumber)
	binary.LittleEndian.PutUint16(buf[4:6], versionMajor)
	binary.LittleEndian.PutUint16(buf[6:8], versionMinor)
	binary.LittleEndian.PutUint32(buf[16:20], snaplen)
	binary.LittleEndian.PutUint32(buf[20:24], uint32(linktype))
	return buf[:]
}

func GetPacketHeader(ci gopacket.CaptureInfo) ([]byte, error) {
	var buf [16]byte
	t := ci.Timestamp
	if t.IsZero() {
		return nil, fmt.Errorf("incoming packet does not have a timestamp. Ignoring packet")
	}
	secs := t.Unix()
	usecs := t.Nanosecond() / nanosPerMicro
	binary.LittleEndian.PutUint32(buf[0:4], uint32(secs))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(usecs))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(ci.CaptureLength))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(ci.Length))
	return buf[:], nil
}

func GetPacketBytesWithHeader(time time.Time, data []byte) ([]byte, error) {
	ci := gopacket.CaptureInfo{
		Timestamp:     time,
		CaptureLength: len(data),
		Length:        len(data),
	}
	if ci.CaptureLength != len(data) {
		return nil, fmt.Errorf("capture length %d does not match data length %d", ci.CaptureLength, len(data))
	}
	if ci.CaptureLength > ci.Length {
		return nil, fmt.Errorf("invalid capture info %+v:  capture length > length", ci)
	}
	b, err := GetPacketHeader(ci)
	if err != nil {
		return nil, fmt.Errorf("error writing packet header: %w", err)
	}
	// append 16 byte packet header & data all at once
	return append(b, data...), nil
}
