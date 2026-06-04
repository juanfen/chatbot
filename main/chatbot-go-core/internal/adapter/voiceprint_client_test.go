package adapter

import "testing"

func TestResolveVoiceprintIdentifyURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		rawURL  string
		apiKey  string
		wantURL string
		wantKey string
	}{
		{
			name:    "identify full path",
			rawURL:  "http://192.168.99.103:8006/voiceprint/identify?key=token1",
			wantURL: "http://192.168.99.103:8006/voiceprint/identify",
			wantKey: "token1",
		},
		{
			name:    "health path",
			rawURL:  "http://127.0.0.1:8005/voiceprint/health?key=token2",
			wantURL: "http://127.0.0.1:8005/voiceprint/identify",
			wantKey: "token2",
		},
		{
			name:    "explicit api key",
			rawURL:  "http://127.0.0.1:8006/voiceprint/identify",
			apiKey:  "token3",
			wantURL: "http://127.0.0.1:8006/voiceprint/identify",
			wantKey: "token3",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotURL, gotKey, err := resolveVoiceprintIdentifyURL(tc.rawURL, tc.apiKey)
			if err != nil {
				t.Fatal(err)
			}
			if gotURL != tc.wantURL {
				t.Fatalf("url: got %q want %q", gotURL, tc.wantURL)
			}
			if gotKey != tc.wantKey {
				t.Fatalf("key: got %q want %q", gotKey, tc.wantKey)
			}
		})
	}
}
