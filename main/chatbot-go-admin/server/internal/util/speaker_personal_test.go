package util

import "testing"

func TestNormalizeSpeakerPersonalLines(t *testing.T) {
	lines, err := NormalizeSpeakerPersonalLines([]string{"  你好  ", "世界"})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != "你好" || lines[1] != "世界" {
		t.Fatalf("got %v", lines)
	}
}

func TestNormalizeSpeakerPersonalLines_TooLong(t *testing.T) {
	long := string([]rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u'})
	_, err := NormalizeSpeakerPersonalLines([]string{long})
	if err == nil {
		t.Fatal("expected error")
	}
}
