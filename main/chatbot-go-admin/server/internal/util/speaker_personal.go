package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	SpeakerPersonalLineMaxRunes = 20
	SpeakerPersonalMaxLines     = 32
)

// ParseSpeakerPersonalLines 从 DB 字段解析个性化条目列表。
func ParseSpeakerPersonalLines(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return []string{}, nil
	}
	var lines []string
	if err := json.Unmarshal([]byte(raw), &lines); err != nil {
		return nil, fmt.Errorf("invalid personal_lines json: %w", err)
	}
	return NormalizeSpeakerPersonalLines(lines)
}

// MarshalSpeakerPersonalLines 序列化为 JSON 存入 DB。
func MarshalSpeakerPersonalLines(lines []string) (string, error) {
	normalized, err := NormalizeSpeakerPersonalLines(lines)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// NormalizeSpeakerPersonalLines 去空白、校验长度与条数上限。
func NormalizeSpeakerPersonalLines(lines []string) ([]string, error) {
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if utf8.RuneCountInString(line) > SpeakerPersonalLineMaxRunes {
			return nil, fmt.Errorf("line %d exceeds %d characters", i+1, SpeakerPersonalLineMaxRunes)
		}
		out = append(out, line)
		if len(out) > SpeakerPersonalMaxLines {
			return nil, fmt.Errorf("at most %d lines allowed", SpeakerPersonalMaxLines)
		}
	}
	return out, nil
}
