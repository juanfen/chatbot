package service

import (
	"strings"
	"testing"
)

func TestSplitTextForTTS_Basic(t *testing.T) {
	// Empty and short text
	if got := splitTextForTTS("", 80); got != nil {
		t.Errorf("empty text should return nil, got %v", got)
	}

	if got := splitTextForTTS("你好", 80); len(got) != 1 || got[0] != "你好" {
		t.Errorf("short text should return single segment, got %v", got)
	}
}

func TestSplitTextForTTS_SentenceBoundary(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		want     []string
	}{
		{
			name:     "split at Chinese period - long text",
			text:     "今天天气真好。我们去公园散步吧。你觉得怎么样？我有很多话想说。",
			maxChars: 20,
			want:     []string{"今天天气真好。", "我们去公园散步吧。", "你觉得怎么样？", "我有很多话想说。"},
		},
		{
			name:     "split at exclamation mark - long text",
			text:     "太棒了！这是一个非常好的主意！我们一起去做吧。",
			maxChars: 15,
			want:     []string{"太棒了！", "这是一个非常好的主意！", "我们一起去做吧。"},
		},
		{
			name:     "split at question mark",
			text:     "你吃饭了吗？我今天做了一道新菜。味道很不错呢。",
			maxChars: 15,
			want:     []string{"你吃饭了吗？", "我今天做了一道新菜。", "味道很不错呢。"},
		},
		{
			name:     "split at semicolon - short items",
			text:     "苹果；香蕉；橙子；葡萄；西瓜",
			maxChars: 8,
			want:     []string{"苹果；", "香蕉；", "橙子；", "葡萄；", "西瓜"},
		},
		{
			name:     "split at comma - short items",
			text:     "我，你，他，大家，所有人",
			maxChars: 6,
			want:     []string{"我，", "你，", "他，", "大家，", "所有人"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitTextForTTS(tt.text, tt.maxChars)
			if len(got) != len(tt.want) {
				t.Errorf("segment count mismatch: want %d, got %d: %v", len(tt.want), len(got), got)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("segment[%d] mismatch:\nwant %q\ngot  %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestSplitTextForTTS_NoPunctuation_Fallback(t *testing.T) {
	// When no punctuation exists, should fallback to fixed split
	text := "这是一个很长很长很长的句子没有标点符号"
	maxChars := 10

	got := splitTextForTTS(text, maxChars)

	// Should split into chunks of ~maxChars
	totalChars := 0
	for _, seg := range got {
		totalChars += len([]rune(seg))
	}
	if totalChars != len([]rune(text)) {
		t.Errorf("total chars mismatch: want %d, got %d", len([]rune(text)), totalChars)
	}

	// Each segment should be <= maxChars (except maybe last)
	for i, seg := range got {
		if i < len(got)-1 && len([]rune(seg)) > maxChars {
			t.Errorf("segment[%d] too long: %d > %d", i, len([]rune(seg)), maxChars)
		}
	}
}

func TestSplitTextForTTS_PunctuationPreservation(t *testing.T) {
	// Punctuation should be preserved at end of segments
	text := "第一句。第二句。第三句。"
	maxChars := 10

	got := splitTextForTTS(text, maxChars)

	// Each segment (except last) should end with punctuation
	for i := 0; i < len(got)-1; i++ {
		seg := strings.TrimSpace(got[i])
		lastChar := []rune(seg)[len([]rune(seg))-1]
		if !containsRune([]rune("。！？.!?；;，,\n"), lastChar) {
			t.Errorf("segment[%d] should end with punctuation, got: %q (last char: %q)",
				i, got[i], string(lastChar))
		}
	}
}

func TestSplitTextForTTS_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		wantLen  int
	}{
		{
			name:     "very long sentence no punctuation",
			text:     "abcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxy",
			maxChars: 20,
			wantLen:  3, // Should split into ~3 segments
		},
		{
			name:     "maxChars larger than text",
			text:     "短句子",
			maxChars: 100,
			wantLen:  1,
		},
		{
			name:     "maxChars zero uses default",
			text:     "测试一下",
			maxChars: 0,
			wantLen:  1,
		},
		{
			name:     "mixed punctuation",
			text:     "你好！今天天气很好。我们出去玩？好的，我们走吧。太好了！",
			maxChars: 15,
			wantLen:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitTextForTTS(tt.text, tt.maxChars)
			if len(got) != tt.wantLen {
				t.Errorf("segment count: want %d, got %d: %v", tt.wantLen, len(got), got)
			}
		})
	}
}

func TestSplitTextForTTS_MultiplePunctuation(t *testing.T) {
	// Multiple punctuation should be handled correctly
	text := "太好了！！！真的吗？？？"
	maxChars := 10

	got := splitTextForTTS(text, maxChars)

	// Should split at first ！ and ？
	if len(got) != 2 {
		t.Errorf("expected 2 segments, got %d: %v", len(got), got)
	}
}

func TestSplitTextForTTS_NewlineBoundary(t *testing.T) {
	text := "第一\n第二\n第三"
	maxChars := 5

	got := splitTextForTTS(text, maxChars)

	// With maxChars=5, each line should fit in one segment
	// Since newline is a soft boundary, it will split when no strong boundary found
	if len(got) < 2 {
		t.Errorf("expected at least 2 segments for newline text, got %d: %v", len(got), got)
	}
}

func TestNormalizeStreamingDelta_SecondPeriodNotDropped(t *testing.T) {
	// 前文已含「。」；句末再单独来一个「。」时，旧 Contains 逻辑会误返回空串。
	existing := "第一句。第二句还没说完"
	incoming := "。"
	if got := normalizeStreamingDelta(existing, incoming); got != "。" {
		t.Fatalf("got %q, want %q", got, "。")
	}
}

func TestNormalizeStreamingDelta_CumulativeStream(t *testing.T) {
	var acc string
	chunks := []string{"你", "你好", "你好。", "你好。世界"}
	for _, ch := range chunks {
		delta := normalizeStreamingDelta(acc, ch)
		acc += delta
	}
	want := "你好。世界"
	if acc != want {
		t.Fatalf("accumulated %q, want %q", acc, want)
	}
}

func TestNormalizeStreamingDelta_ExactDuplicate(t *testing.T) {
	if got := normalizeStreamingDelta("same", "same"); got != "" {
		t.Fatalf("want empty duplicate, got %q", got)
	}
}
