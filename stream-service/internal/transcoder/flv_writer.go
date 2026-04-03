package transcoder

import (
	"bytes"
	"encoding/binary"
	"io"
)

// FLVWriter writes FLV (Flash Video) format data
// This is used to convert RTMP data to FLV format that FFmpeg can read from stdin
type FLVWriter struct {
	writer      io.Writer
	headerWrote bool
}

// NewFLVWriter creates a new FLV writer
func NewFLVWriter(writer io.Writer) *FLVWriter {
	return &FLVWriter{
		writer:      writer,
		headerWrote: false,
	}
}

// WriteFLVHeader writes the FLV file header
func (w *FLVWriter) WriteFLVHeader() error {
	if w.headerWrote {
		return nil
	}

	// FLV header format:
	// Signature: "FLV" (3 bytes)
	// Version: 1 (1 byte)
	// Flags: 0x05 (audio + video) (1 byte)
	// Header size: 9 (4 bytes)
	// Previous tag size: 0 (4 bytes)
	header := []byte{
		'F', 'L', 'V',  // Signature
		0x01,           // Version 1
		0x05,           // Flags: audio + video (0x04 = audio, 0x01 = video, 0x05 = both)
		0x00, 0x00, 0x00, 0x09, // Header size (9 bytes)
		0x00, 0x00, 0x00, 0x00, // Previous tag size 0 (first tag)
	}

	_, err := w.writer.Write(header)
	if err != nil {
		return err
	}

	w.headerWrote = true
	return nil
}

// WriteTag writes an FLV tag (audio, video, or script data)
func (w *FLVWriter) WriteTag(tagType byte, timestamp uint32, data []byte) error {
	if !w.headerWrote {
		if err := w.WriteFLVHeader(); err != nil {
			return err
		}
	}

	dataSize := len(data)

	// FLV tag format:
	// Tag type: 1 byte (8=audio, 9=video, 18=script data)
	// Data size: 3 bytes
	// Timestamp: 3 bytes (lower 24 bits)
	// Timestamp extended: 1 byte (upper 8 bits)
	// Stream ID: 3 bytes (always 0)
	// Data: N bytes
	// Previous tag size: 4 bytes (tag header size + data size)

	var buf bytes.Buffer

	// Tag type
	buf.WriteByte(tagType)

	// Data size (24 bits, big-endian)
	buf.WriteByte(byte((dataSize >> 16) & 0xFF))
	buf.WriteByte(byte((dataSize >> 8) & 0xFF))
	buf.WriteByte(byte(dataSize & 0xFF))

	// Timestamp (24 bits, big-endian)
	buf.WriteByte(byte((timestamp >> 16) & 0xFF))
	buf.WriteByte(byte((timestamp >> 8) & 0xFF))
	buf.WriteByte(byte(timestamp & 0xFF))

	// Timestamp extended (8 bits) - upper 8 bits of 32-bit timestamp
	buf.WriteByte(byte((timestamp >> 24) & 0xFF))

	// Stream ID (always 0)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)

	// Write tag header
	if _, err := w.writer.Write(buf.Bytes()); err != nil {
		return err
	}

	// Write data
	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	// Write previous tag size (tag header size + data size = 11 + dataSize)
	previousTagSize := uint32(11 + dataSize)
	if err := binary.Write(w.writer, binary.BigEndian, previousTagSize); err != nil {
		return err
	}

	return nil
}

// WriteAudioTag writes an audio tag
func (w *FLVWriter) WriteAudioTag(timestamp uint32, data []byte) error {
	return w.WriteTag(0x08, timestamp, data)
}

// WriteVideoTag writes a video tag
func (w *FLVWriter) WriteVideoTag(timestamp uint32, data []byte) error {
	return w.WriteTag(0x09, timestamp, data)
}

// WriteScriptTag writes a script/metadata tag
func (w *FLVWriter) WriteScriptTag(timestamp uint32, data []byte) error {
	return w.WriteTag(0x12, timestamp, data)
}
