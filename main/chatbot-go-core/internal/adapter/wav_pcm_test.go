package adapter

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

// buildPCM16MonoWAV 构造最小合法 WAV（PCM s16le mono），pcm 为 data 子块负载。
func buildPCM16MonoWAV(sampleRate int, pcm []byte) []byte {
	const bitsPerSample = 16
	const channels = 1
	const audioFormat = 1
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign

	fmtChunk := make([]byte, 4+4+16)
	copy(fmtChunk[0:4], []byte("fmt "))
	binary.LittleEndian.PutUint32(fmtChunk[4:8], 16)
	binary.LittleEndian.PutUint16(fmtChunk[8:10], audioFormat)
	binary.LittleEndian.PutUint16(fmtChunk[10:12], channels)
	binary.LittleEndian.PutUint32(fmtChunk[12:16], uint32(sampleRate))
	binary.LittleEndian.PutUint32(fmtChunk[16:20], uint32(byteRate))
	binary.LittleEndian.PutUint16(fmtChunk[20:22], uint16(blockAlign))
	binary.LittleEndian.PutUint16(fmtChunk[22:24], bitsPerSample)

	dataChunk := make([]byte, 8+len(pcm))
	copy(dataChunk[0:4], []byte("data"))
	binary.LittleEndian.PutUint32(dataChunk[4:8], uint32(len(pcm)))
	copy(dataChunk[8:], pcm)

	wave := append(fmtChunk, dataChunk...)
	out := make([]byte, 4+4+4+len(wave))
	copy(out[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(out[4:8], uint32(4+len(wave)))
	copy(out[8:12], []byte("WAVE"))
	copy(out[12:], wave)
	return out
}

func TestMaybePCMReaderFromHTTPBody_WAVStripsHeader(t *testing.T) {
	pcm := bytes.Repeat([]byte{0x01, 0x02}, 1440) // 2880 bytes = one 60ms frame @ 24k mono s16le
	wav := buildPCM16MonoWAV(24000, pcm)
	r, stripped, err := maybePCMReaderFromHTTPBody("audio/wav", bytes.NewReader(wav))
	if err != nil {
		t.Fatal(err)
	}
	if !stripped {
		t.Fatal("expected wav stripped")
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pcm) {
		t.Fatalf("pcm mismatch: want len %d got len %d", len(pcm), len(got))
	}
}

func TestMaybePCMReaderFromHTTPBody_RawPCM(t *testing.T) {
	raw := []byte{1, 2, 3, 4, 5, 6}
	r, stripped, err := maybePCMReaderFromHTTPBody("application/octet-stream", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if stripped {
		t.Fatal("expected not wav")
	}
	got, _ := io.ReadAll(r)
	if !bytes.Equal(got, raw) {
		t.Fatalf("got %v want %v", got, raw)
	}
}
