package adapter

import (
	"encoding/binary"
	"fmt"

	"gopkg.in/hraban/opus.v2"
)

type OpusEncoder struct {
	encoder    *opus.Encoder
	sampleRate int
	channels   int
}

func NewOpusEncoder(sampleRate int, channels int) (*OpusEncoder, error) {
	if sampleRate <= 0 {
		sampleRate = 24000
	}
	if channels <= 0 {
		channels = 1
	}
	encoder, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("create opus encoder: %w", err)
	}
	return &OpusEncoder{encoder: encoder, sampleRate: sampleRate, channels: channels}, nil
}

func (e *OpusEncoder) SampleRate() int { return e.sampleRate }
func (e *OpusEncoder) Channels() int   { return e.channels }

// PCMS16FrameBytes 返回固定帧长对应的 16-bit LE PCM 字节数。
func PCMS16FrameBytes(sampleRate, channels, frameMs int) int {
	if sampleRate <= 0 {
		sampleRate = 24000
	}
	if channels <= 0 {
		channels = 1
	}
	if frameMs <= 0 {
		frameMs = 60
	}
	return sampleRate * channels * frameMs / 1000 * 2
}

// PCMS16DurationMs 根据 PCM 字节数估算播放时长（毫秒）。
func PCMS16DurationMs(pcmBytes, sampleRate, channels int) int {
	if pcmBytes <= 0 || sampleRate <= 0 || channels <= 0 {
		return 0
	}
	samples := pcmBytes / 2
	return samples * 1000 / (sampleRate * channels)
}

// EncodePCMS16LEToOpusFrames 将 PCM 按 frameMs 切帧后逐帧 Opus 编码（末帧不足则补零）。
func EncodePCMS16LEToOpusFrames(pcm []byte, sampleRate, channels, frameMs int) ([][]byte, error) {
	enc, err := NewOpusEncoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}
	return enc.EncodeS16LEFrames(pcm, frameMs)
}

// EncodeS16LE 将整段 PCM 一次编码为一个 Opus 包（仅用于短音频；TTS 下行请用 EncodeS16LEFrames）。
func (e *OpusEncoder) EncodeS16LE(pcm []byte) ([]byte, error) {
	frames, err := e.EncodeS16LEFrames(pcm, 60)
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, nil
	}
	if len(frames) == 1 {
		return frames[0], nil
	}
	var out []byte
	for _, f := range frames {
		out = append(out, f...)
	}
	return out, nil
}

// EncodeS16LEFrames 按固定帧长切分 PCM，每帧单独 Opus 编码，保证每包对应 frameMs 时长。
func (e *OpusEncoder) EncodeS16LEFrames(pcm []byte, frameMs int) ([][]byte, error) {
	if len(pcm) == 0 {
		return nil, nil
	}
	if frameMs <= 0 {
		frameMs = 60
	}
	frameBytes := PCMS16FrameBytes(e.sampleRate, e.channels, frameMs)
	if frameBytes <= 0 {
		return nil, fmt.Errorf("invalid opus frame: sampleRate=%d channels=%d frameMs=%d", e.sampleRate, e.channels, frameMs)
	}
	if len(pcm)%2 != 0 {
		pcm = pcm[:len(pcm)-1]
	}

	var out [][]byte
	for len(pcm) > 0 {
		chunk := make([]byte, frameBytes)
		n := copy(chunk, pcm)
		pcm = pcm[n:]
		packet, err := e.encodePCMFrame(chunk)
		if err != nil {
			return nil, err
		}
		if len(packet) > 0 {
			out = append(out, packet)
		}
	}
	return out, nil
}

func (e *OpusEncoder) encodePCMFrame(pcmFrame []byte) ([]byte, error) {
	if len(pcmFrame)%2 != 0 {
		pcmFrame = pcmFrame[:len(pcmFrame)-1]
	}
	samples := make([]int16, len(pcmFrame)/2)
	for i := range samples {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmFrame[i*2:]))
	}
	out := make([]byte, 4096)
	n, err := e.encoder.Encode(samples, out)
	if err != nil {
		return nil, fmt.Errorf("encode opus: %w", err)
	}
	return append([]byte(nil), out[:n]...), nil
}
