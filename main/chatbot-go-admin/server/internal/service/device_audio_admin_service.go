package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ListManageableDeviceAudio lists device audio rows for the admin UI.
func (s *DeviceAudioService) ListManageableDeviceAudio(ctx context.Context, userID string, page, pageSize int) (*dto.DeviceAudioAdminListResp, error) {
	rows, total, err := s.repo.ListManageable(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	speakerNameByID := map[string]string{}
	if s.speakerRepo != nil && userID != "" {
		speakers, err := s.speakerRepo.ListByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, sp := range speakers {
			speakerNameByID[sp.SpeakerID] = sp.SpeakerName
		}
	}

	list := make([]dto.DeviceAudioAdminItem, 0, len(rows))
	for _, r := range rows {
		item := dto.DeviceAudioAdminItem{
			ID:         r.ID,
			MacID:      r.MacID,
			AudioLen:   r.AudioLen,
			Context:    r.Context,
			CreateTime: r.CreateTime,
			ExpireTime: r.ExpireTime,
			DeleteTime: r.DeleteTime,
			Perpetual:  util.IsPerpetual(r.ExpireTime),
		}
		if r.SpeakerID != nil && *r.SpeakerID != "" {
			item.Bound = true
			item.SpeakerID = *r.SpeakerID
			if r.BoundUserID != nil && *r.BoundUserID == userID {
				item.SpeakerName = speakerNameByID[*r.SpeakerID]
			}
		}
		list = append(list, item)
	}

	return &dto.DeviceAudioAdminListResp{Total: total, List: list}, nil
}

// ListUserSpeakers returns speakers created by the user.
func (s *DeviceAudioService) ListUserSpeakers(ctx context.Context, userID string) ([]dto.UserSpeakerItem, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	rows, err := s.speakerRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.UserSpeakerItem, 0, len(rows))
	for _, r := range rows {
		lines, _ := util.ParseSpeakerPersonalLines(r.PersonalLines)
		out = append(out, dto.UserSpeakerItem{
			SpeakerID:     r.SpeakerID,
			SpeakerName:   r.SpeakerName,
			PersonalLines: lines,
			CreateTime:    r.CreateTime,
		})
	}
	return out, nil
}

// GetSpeakerPersonalLines returns personalization lines for one speaker owned by the user.
func (s *DeviceAudioService) GetSpeakerPersonalLines(ctx context.Context, userID, speakerID string) (*dto.SpeakerPersonalLinesResp, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	speakerID = strings.TrimSpace(speakerID)
	if speakerID == "" {
		return nil, ErrSpeakerIDRequired
	}
	row, err := s.speakerRepo.FindByIDAndUser(ctx, speakerID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpeakerNotFound
		}
		return nil, err
	}
	lines, err := util.ParseSpeakerPersonalLines(row.PersonalLines)
	if err != nil {
		return nil, ErrSpeakerPersonalInvalid
	}
	return &dto.SpeakerPersonalLinesResp{
		SpeakerID:     row.SpeakerID,
		SpeakerName:   row.SpeakerName,
		PersonalLines: lines,
	}, nil
}

// SetSpeakerPersonalLines replaces all personalization lines for a speaker.
func (s *DeviceAudioService) SetSpeakerPersonalLines(ctx context.Context, userID, speakerID string, lines []string) (*dto.SpeakerPersonalLinesResp, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	speakerID = strings.TrimSpace(speakerID)
	if speakerID == "" {
		return nil, ErrSpeakerIDRequired
	}
	normalized, err := util.NormalizeSpeakerPersonalLines(lines)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpeakerPersonalInvalid, err)
	}
	jsonText, err := util.MarshalSpeakerPersonalLines(normalized)
	if err != nil {
		return nil, ErrSpeakerPersonalInvalid
	}
	row, err := s.speakerRepo.FindByIDAndUser(ctx, speakerID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpeakerNotFound
		}
		return nil, err
	}
	if err := s.speakerRepo.UpdatePersonalLines(ctx, speakerID, jsonText); err != nil {
		return nil, err
	}
	if len(normalized) > 0 {
		if err := s.syncPersonalLinesToProfile(ctx, speakerID, row.SpeakerName, normalized); err != nil {
			return nil, err
		}
	}
	return &dto.SpeakerPersonalLinesResp{
		SpeakerID:     row.SpeakerID,
		SpeakerName:   row.SpeakerName,
		PersonalLines: normalized,
	}, nil
}

func (s *DeviceAudioService) syncPersonalLinesToProfile(
	ctx context.Context,
	speakerID, speakerName string,
	lines []string,
) error {
	if s.profileImport == nil || !s.profileImport.Enabled() {
		s.logger.Info("【画像导入】未配置 MEMORY_PROFILE_BASE_URL，跳过同步",
			zap.String("音色ID", speakerID),
			zap.String("音色名称", speakerName),
		)
		return nil
	}
	macIDs, err := s.repo.ListMacIDsBySpeakerID(ctx, speakerID)
	if err != nil {
		return err
	}
	if len(macIDs) == 0 {
		s.logger.Warn("【画像导入】音色未绑定设备，无法同步",
			zap.String("音色ID", speakerID),
			zap.String("音色名称", speakerName),
		)
		return ErrSpeakerMacNotFound
	}

	s.logger.Info("【画像导入】开始同步音色个性化内容到记忆服务",
		zap.String("音色ID", speakerID),
		zap.String("音色名称", speakerName),
		zap.Strings("绑定设备MAC列表", macIDs),
		zap.Strings("个性化条目", lines),
		zap.Int("条目数", len(lines)),
	)

	metadata := map[string]any{
		"speaker_id":   speakerID,
		"speaker_name": speakerName,
	}
	for i, macID := range macIDs {
		s.logger.Info("【画像导入】正在同步第 N 个设备",
			zap.Int("当前序号", i+1),
			zap.Int("设备总数", len(macIDs)),
			zap.String("设备MAC", macID),
			zap.String("音色ID", speakerID),
		)
		if _, err := s.profileImport.ImportBasicInfo(ctx, macID, speakerID, lines, metadata); err != nil {
			s.logger.Warn("【画像导入】设备同步失败",
				zap.String("音色ID", speakerID),
				zap.String("设备MAC", macID),
				zap.Error(err),
			)
			return fmt.Errorf("%w: %v", ErrProfileImportFailed, err)
		}
	}
	s.logger.Info("【画像导入】全部设备同步完成",
		zap.String("音色ID", speakerID),
		zap.String("音色名称", speakerName),
		zap.Int("同步设备数", len(macIDs)),
	)
	return nil
}

// BindDeviceAudioSpeaker binds one device audio row to a user speaker.
func (s *DeviceAudioService) BindDeviceAudioSpeaker(ctx context.Context, userID string, req dto.DeviceAudioBindReq) (*dto.DeviceAudioBindResp, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	req.DeviceAudioID = strings.TrimSpace(req.DeviceAudioID)
	req.SpeakerID = strings.TrimSpace(req.SpeakerID)
	req.SpeakerName = strings.TrimSpace(req.SpeakerName)
	if req.DeviceAudioID == "" {
		return nil, ErrDeviceAudioIDRequired
	}

	audio, err := s.repo.FindByID(ctx, req.DeviceAudioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeviceAudioNotFound
		}
		return nil, err
	}
	if audio.SpeakerID != nil && *audio.SpeakerID != "" {
		return nil, ErrDeviceAudioAlreadyBound
	}
	if audio.ExpireTime.Before(time.Now()) && !util.IsPerpetual(audio.ExpireTime) {
		return nil, ErrDeviceAudioExpired
	}

	var speaker *model.BizUserSpeaker
	switch {
	case req.SpeakerID != "":
		speaker, err = s.speakerRepo.FindByIDAndUser(ctx, req.SpeakerID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrSpeakerNotFound
			}
			return nil, err
		}
	case req.SpeakerName != "":
		speaker, err = s.speakerRepo.FindByNameAndUser(ctx, userID, req.SpeakerName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				now := time.Now()
				speaker = &model.BizUserSpeaker{
					SpeakerID:     util.NewSnowflakeID32(),
					UserID:        userID,
					SpeakerName:   req.SpeakerName,
					PersonalLines: "[]",
					CreateTime:    now,
					UpdateTime:    now,
				}
				if err := s.speakerRepo.Create(ctx, speaker); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	default:
		return nil, ErrSpeakerNameOrIDRequired
	}

	if len(audio.AudioBlob) == 0 {
		return nil, ErrDeviceAudioEmpty
	}
	if err := s.reportVoiceprint(ctx, req.DeviceAudioID, speaker.SpeakerID, speaker.SpeakerName, audio.AudioBlob); err != nil {
		return nil, err
	}

	if err := s.repo.BindSpeaker(ctx, req.DeviceAudioID, speaker.SpeakerID, userID); err != nil {
		return nil, err
	}
	_ = s.speakerRepo.TouchUpdateTime(ctx, speaker.SpeakerID)

	return &dto.DeviceAudioBindResp{
		DeviceAudioID: req.DeviceAudioID,
		SpeakerID:     speaker.SpeakerID,
		SpeakerName:   speaker.SpeakerName,
	}, nil
}

// GetDeviceAudioBlob returns WAV bytes for playback.
func (s *DeviceAudioService) GetDeviceAudioBlob(ctx context.Context, audioID string) ([]byte, error) {
	audioID = strings.TrimSpace(audioID)
	if audioID == "" {
		return nil, ErrDeviceAudioIDRequired
	}
	audio, err := s.repo.FindByID(ctx, audioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeviceAudioNotFound
		}
		return nil, err
	}
	if audio.ExpireTime.Before(time.Now()) && !util.IsPerpetual(audio.ExpireTime) {
		return nil, ErrDeviceAudioExpired
	}
	if len(audio.AudioBlob) == 0 {
		return nil, ErrDeviceAudioEmpty
	}
	return audio.AudioBlob, nil
}

func (s *DeviceAudioService) reportVoiceprint(
	ctx context.Context,
	deviceAudioID, speakerID, speakerName string,
	wav []byte,
) error {
	if s.voiceprint == nil || !s.voiceprint.Enabled() {
		s.logger.Warn("sv voiceprint skipped: not configured",
			zap.String("device_audio_id", deviceAudioID),
			zap.String("speaker_id", speakerID),
		)
		return ErrVoiceprintNotConfigured
	}

	s.logger.Info("sv voiceprint bind: start report",
		zap.String("device_audio_id", deviceAudioID),
		zap.String("speaker_id", speakerID),
		zap.String("speaker_name", speakerName),
		zap.Int("audio_bytes", len(wav)),
	)

	result, err := s.voiceprint.Register(ctx, speakerID, wav)
	if err != nil {
		s.logger.Warn("sv voiceprint bind: report failed, abort db bind",
			zap.String("device_audio_id", deviceAudioID),
			zap.String("speaker_id", speakerID),
			zap.String("speaker_name", speakerName),
			zap.Error(err),
		)
		return fmt.Errorf("%w: %v", ErrVoiceprintReportFailed, err)
	}

	s.logger.Info("sv voiceprint bind: report ok, continue db bind",
		zap.String("device_audio_id", deviceAudioID),
		zap.String("speaker_id", speakerID),
		zap.String("speaker_name", speakerName),
		zap.String("response_speaker_id", result.SpeakerID),
		zap.Float64("score", result.Score),
		zap.String("message", result.Message),
	)
	return nil
}
