package adapter

import (
	"encoding/binary"
	"fmt"

	"gopkg.in/hraban/opus.v2"
)

type OpusDecoder struct {
	decoder *opus.Decoder
}

func NewOpusDecoder(sampleRate int, channels int) (*OpusDecoder, error) {
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	if channels <= 0 {
		channels = 1
	}
	decoder, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, fmt.Errorf("create opus decoder: %w", err)
	}
	return &OpusDecoder{decoder: decoder}, nil
}

func (d *OpusDecoder) Decode(packet []byte, frameSize int) ([]byte, error) {
	if len(packet) == 0 {
		return nil, nil
	}
	if frameSize <= 0 {
		frameSize = 960
	}
	pcm := make([]int16, frameSize)
	n, err := d.decoder.Decode(packet, pcm)
	if err != nil {
		return nil, fmt.Errorf("decode opus packet: %w", err)
	}
	out := make([]byte, n*2)
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(pcm[i]))
	}
	return out, nil
}
