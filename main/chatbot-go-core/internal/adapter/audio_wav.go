package adapter

import "bytes"

func PCM16ToWAV(pcm []byte, sampleRate int, channels int) []byte {
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	if channels <= 0 {
		channels = 1
	}
	dataSize := uint32(len(pcm))
	byteRate := uint32(sampleRate * channels * 2)
	blockAlign := uint16(channels * 2)
	var b bytes.Buffer
	b.WriteString("RIFF")
	writeLE32(&b, 36+dataSize)
	b.WriteString("WAVEfmt ")
	writeLE32(&b, 16)
	writeLE16(&b, 1)
	writeLE16(&b, uint16(channels))
	writeLE32(&b, uint32(sampleRate))
	writeLE32(&b, byteRate)
	writeLE16(&b, blockAlign)
	writeLE16(&b, 16)
	b.WriteString("data")
	writeLE32(&b, dataSize)
	b.Write(pcm)
	return b.Bytes()
}

func writeLE16(b *bytes.Buffer, v uint16) {
	b.WriteByte(byte(v))
	b.WriteByte(byte(v >> 8))
}

func writeLE32(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 24))
}
