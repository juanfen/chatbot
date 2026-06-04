package service

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrCannotDeleteSelf  = errors.New("cannot delete yourself")
	ErrLastAdminUser     = errors.New("cannot delete the last admin user")
	ErrRoleInUse         = errors.New("role is assigned to users")
	ErrReservedRole      = errors.New("reserved role cannot be deleted")
	ErrRouteNameExists   = errors.New("route name already exists")
	ErrRouteHasChildren  = errors.New("route has child routes")
	ErrCannotDeleteHome  = errors.New("cannot delete home route")

	ErrPackageNotFound      = errors.New("package not found")
	ErrPackageDisabled      = errors.New("package disabled")
	ErrInsufficientQuota    = errors.New("insufficient quota")
	ErrUserQuotaNotFound    = errors.New("user quota not configured")

	ErrDeviceMacIDRequired  = errors.New("mac_id is required")
	ErrDeviceAudioLenRequired = errors.New("audio_len is required")
	ErrDeviceAudioEmpty     = errors.New("audio_blob is empty")
	ErrDeviceAudioTooLarge  = errors.New("audio_blob exceeds size limit")
	ErrDeviceAudioNotWAV    = errors.New("audio_blob must be wav")

	ErrUnauthorized              = errors.New("unauthorized")
	ErrDeviceAudioIDRequired     = errors.New("deviceAudioId is required")
	ErrDeviceAudioNotFound       = errors.New("device audio not found")
	ErrDeviceAudioAlreadyBound   = errors.New("device audio already bound to a speaker")
	ErrDeviceAudioExpired        = errors.New("device audio expired")
	ErrSpeakerIDRequired         = errors.New("speaker_id is required")
	ErrSpeakerNotFound           = errors.New("speaker not found")
	ErrSpeakerNameOrIDRequired   = errors.New("speakerId or speakerName is required")
	ErrSpeakerPersonalInvalid    = errors.New("invalid speaker personal lines")

	ErrVoiceprintNotConfigured = errors.New("voiceprint service not configured")
	ErrVoiceprintReportFailed  = errors.New("voiceprint report failed")

	ErrSpeakerMacNotFound      = errors.New("该音色尚未绑定设备，请先分配音色后再保存个性化内容")
	ErrProfileImportFailed     = errors.New("profile import failed")
)
