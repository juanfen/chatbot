package dto

import "time"

// DeviceAudioUploadResp is returned after a successful upload.
type DeviceAudioUploadResp struct {
	ID         string    `json:"id"`
	MacID      string    `json:"macId"`
	AudioLen   string    `json:"audioLen"`
	Context    string    `json:"context"`
	CreateTime time.Time `json:"createTime"`
	ExpireTime time.Time `json:"expireTime"`
	DeleteTime time.Time `json:"deleteTime"`
}

// DeviceSpeakerIDsResp lists registered speaker IDs for a device.
type DeviceSpeakerIDsResp struct {
	MacID      string   `json:"macId"`
	SpeakerIDs []string `json:"speakerIds"`
}

// DeviceSpeakerNameResp is the speaker name lookup result.
type DeviceSpeakerNameResp struct {
	SpeakerID     string   `json:"speakerId"`
	SpeakerName   string   `json:"speakerName"`
	PersonalLines []string `json:"personalLines"`
}

// DeviceAudioListItem is a query result without binary payload.
type DeviceAudioListItem struct {
	ID         string    `json:"id"`
	MacID      string    `json:"macId"`
	AudioLen   string    `json:"audioLen"`
	Context    string    `json:"context"`
	CreateTime time.Time `json:"createTime"`
	ExpireTime time.Time `json:"expireTime"`
}
