package adapter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// maybePCMReaderFromHTTPBody 若响应为 WAV（RIFF/WAVE），跳过 fmt 等子块直到 data 子块，
// 仅对 data 中的原始 PCM 做后续分帧；否则原样返回（兼容纯 PCM 流）。
// 第二个返回值表示是否识别并剥离了 WAV 容器（便于日志）。
func maybePCMReaderFromHTTPBody(contentType string, body io.Reader) (io.Reader, bool, error) {
	br := bufio.NewReaderSize(body, 64*1024)
	head, err := br.Peek(12)
	if err != nil && err != io.EOF {
		return nil, false, err
	}
	if len(head) < 12 || string(head[0:4]) != "RIFF" || string(head[8:12]) != "WAVE" {
		// 部分网关误标 audio/wav，实为裸 PCM；按裸流处理。
		return br, false, nil
	}
	if _, err := io.ReadFull(br, make([]byte, 12)); err != nil {
		return nil, false, fmt.Errorf("wav: read riff header: %w", err)
	}
	for i := 0; i < 256; i++ {
		var chunkID [4]byte
		if _, err := io.ReadFull(br, chunkID[:]); err != nil {
			return nil, false, fmt.Errorf("wav: read chunk id: %w", err)
		}
		var chunkSize uint32
		if err := binary.Read(br, binary.LittleEndian, &chunkSize); err != nil {
			return nil, false, fmt.Errorf("wav: read chunk size: %w", err)
		}
		if string(chunkID[:]) == "data" {
			if chunkSize == 0 {
				return br, true, nil
			}
			return io.LimitReader(br, int64(chunkSize)), true, nil
		}
		skip := int64(chunkSize)
		if skip < 0 {
			return nil, false, fmt.Errorf("wav: invalid chunk size %d", chunkSize)
		}
		if chunkSize%2 == 1 {
			skip++
		}
		if _, err := io.CopyN(io.Discard, br, skip); err != nil {
			return nil, false, fmt.Errorf("wav: skip chunk %q: %w", string(chunkID[:]), err)
		}
	}
	return nil, false, fmt.Errorf("wav: no data chunk in first %d subchunks", 256)
}
