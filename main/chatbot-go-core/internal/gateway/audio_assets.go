package gateway

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"xiaozhi-go-core/internal/adapter"
)

func loadAudioFileAsOpus(path string, sampleRate int, channels int, frameMs int) ([][]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	pcm, err := audioFileToPCM(path, sampleRate, channels)
	if err != nil {
		return nil, err
	}
	return pcmToOpusFrames(pcm, sampleRate, channels, frameMs)
}

func audioFileToPCM(path string, sampleRate int, channels int) ([]byte, error) {
	if strings.EqualFold(filepath.Ext(path), ".wav") {
		if pcm, err := readWAVPCM(path); err == nil {
			return pcm, nil
		}
	}
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", path,
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", fmt.Sprintf("%d", channels),
		"-ar", fmt.Sprintf("%d", sampleRate),
		"pipe:1",
	}
	cmd := exec.Command("ffmpeg", args...)
	return cmd.Output()
}

func readWAVPCM(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not wav")
	}
	idx := bytes.Index(data, []byte("data"))
	if idx < 0 || idx+8 > len(data) {
		return nil, fmt.Errorf("missing data chunk")
	}
	size := int(binary.LittleEndian.Uint32(data[idx+4 : idx+8]))
	start := idx + 8
	if start+size > len(data) {
		size = len(data) - start
	}
	return append([]byte(nil), data[start:start+size]...), nil
}

func pcmToOpusFrames(pcm []byte, sampleRate int, channels int, frameMs int) ([][]byte, error) {
	return adapter.EncodePCMS16LEToOpusFrames(pcm, sampleRate, channels, frameMs)
}

func readWakeupResponse(configFile string, fallbackFile string) (string, string) {
	data, err := os.ReadFile(configFile)
	if err == nil {
		var raw map[string]map[string]any
		if json.Unmarshal(data, &raw) == nil {
			for _, item := range raw {
				if path, ok := item["file_path"].(string); ok {
					text, _ := item["text"].(string)
					if stat, err := os.Stat(path); err == nil && stat.Size() >= 15*1024 {
						return path, text
					}
				}
			}
		}
		lines := strings.Split(string(data), "\n")
		var path, text string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "file_path:") {
				path = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "file_path:")), `"'`)
			}
			if strings.HasPrefix(line, "text:") {
				text = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "text:")), `"'`)
			}
		}
		if path != "" {
			if stat, err := os.Stat(path); err == nil && stat.Size() >= 15*1024 {
				if text == "" {
					text = "我在这里哦！"
				}
				return path, text
			}
		}
	}
	return fallbackFile, "我在这里哦！"
}
