package dto

import "time"

// DeviceAudioAdminItem is a row in the voice recognition admin table.
type DeviceAudioAdminItem struct {
	ID          string     `json:"id"`
	MacID       string     `json:"macId"`
	AudioLen    string     `json:"audioLen"`
	Context     string     `json:"context"`
	SpeakerID   string     `json:"speakerId,omitempty"`
	SpeakerName string     `json:"speakerName,omitempty"`
	Bound       bool       `json:"bound"`
	Perpetual   bool       `json:"perpetual"`
	CreateTime  time.Time  `json:"createTime"`
	ExpireTime  time.Time  `json:"expireTime"`
	DeleteTime  time.Time  `json:"deleteTime"`
}

// DeviceAudioAdminListResp paginated list.
type DeviceAudioAdminListResp struct {
	Total int64                  `json:"total"`
	List  []DeviceAudioAdminItem `json:"list"`
}

// UserSpeakerItem is a speaker owned by the current user.
type UserSpeakerItem struct {
	SpeakerID     string    `json:"speakerId"`
	SpeakerName   string    `json:"speakerName"`
	PersonalLines []string  `json:"personalLines"`
	CreateTime    time.Time `json:"createTime"`
}

// SpeakerPersonalLinesReq 保存某音色的全部个性化条目（覆盖式）。
type SpeakerPersonalLinesReq struct {
	Lines []string `json:"lines"`
}

// SpeakerPersonalLinesResp 查询/保存后的个性化条目。
type SpeakerPersonalLinesResp struct {
	SpeakerID     string   `json:"speakerId"`
	SpeakerName   string   `json:"speakerName"`
	PersonalLines []string `json:"personalLines"`
}

// DeviceAudioBindReq binds a device audio row to a speaker.
type DeviceAudioBindReq struct {
	DeviceAudioID string `json:"deviceAudioId" binding:"required"`
	SpeakerID     string `json:"speakerId"`
	SpeakerName   string `json:"speakerName"`
}

// DeviceAudioBindResp result after bind.
type DeviceAudioBindResp struct {
	DeviceAudioID string `json:"deviceAudioId"`
	SpeakerID     string `json:"speakerId"`
	SpeakerName   string `json:"speakerName"`
}
