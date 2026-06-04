package dto

import "time"

// MobileVoiceItem is one stored voice sample for the mobile UI.
type MobileVoiceItem struct {
	ID              string    `json:"id"`
	FileName        string    `json:"fileName"`
	UploadTime      time.Time `json:"uploadTime"`
	DurationSeconds float64   `json:"durationSeconds"`
	IsActive        bool      `json:"isActive"`
}

// MobileStandardTextResp wraps standard reading text.
type MobileStandardTextResp struct {
	StandardText string `json:"standardText"`
}

// MobileStandardTextReq updates standard reading text.
type MobileStandardTextReq struct {
	StandardText string `json:"standardText"`
}
