package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"chatbot-go-admin/server/internal/adapter"
	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"chatbot-go-admin/server/internal/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	deviceAudioExpireAfter   = 7 * 24 * time.Hour
	deviceAudioDeleteAfter   = 14 * 24 * time.Hour
	deviceAudioMaxBytes      = 32 << 20 // 32 MiB
	deviceAudioDefaultContext = "暂无内容"
)

type DeviceAudioService struct {
	repo          *repository.BizDeviceAudioRepository
	speakerRepo   *repository.BizUserSpeakerRepository
	voiceprint    *adapter.VoiceprintClient
	profileImport *adapter.ProfileImportClient
	logger        *zap.Logger
}

func NewDeviceAudioService(
	repo *repository.BizDeviceAudioRepository,
	speakerRepo *repository.BizUserSpeakerRepository,
	voiceprint *adapter.VoiceprintClient,
	profileImport *adapter.ProfileImportClient,
	logger *zap.Logger,
) *DeviceAudioService {
	return &DeviceAudioService{
		repo:          repo,
		speakerRepo:   speakerRepo,
		voiceprint:    voiceprint,
		profileImport: profileImport,
		logger:        logger,
	}
}

// Upload stores device WAV audio and lifecycle timestamps.
func (s *DeviceAudioService) Upload(ctx context.Context, macID, audioLen, textContext string, wavData []byte, filename string) (*dto.DeviceAudioUploadResp, error) {
	macID = strings.TrimSpace(macID)
	audioLen = strings.TrimSpace(audioLen)
	textContext = strings.TrimSpace(textContext)
	if textContext == "" {
		textContext = deviceAudioDefaultContext
	}
	if macID == "" {
		return nil, ErrDeviceMacIDRequired
	}
	if audioLen == "" {
		return nil, ErrDeviceAudioLenRequired
	}
	if len(wavData) == 0 {
		return nil, ErrDeviceAudioEmpty
	}
	if len(wavData) > deviceAudioMaxBytes {
		return nil, ErrDeviceAudioTooLarge
	}
	if !isWAVPayload(wavData, filename) {
		return nil, ErrDeviceAudioNotWAV
	}

	now := time.Now()
	row := &model.BizDeviceAudio{
		ID:         uuid.NewString(),
		MacID:      macID,
		AudioBlob:  wavData,
		AudioLen:   audioLen,
		Context:    textContext,
		CreateTime: now,
		ExpireTime: now.Add(deviceAudioExpireAfter),
		DeleteTime: now.Add(deviceAudioDeleteAfter),
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}

	return &dto.DeviceAudioUploadResp{
		ID:         row.ID,
		MacID:      row.MacID,
		AudioLen:   row.AudioLen,
		Context:    row.Context,
		CreateTime: row.CreateTime,
		ExpireTime: row.ExpireTime,
		DeleteTime: row.DeleteTime,
	}, nil
}

// ListSpeakerIDsByMacID returns distinct speaker_id values bound for the device.
func (s *DeviceAudioService) ListSpeakerIDsByMacID(ctx context.Context, macID string) (*dto.DeviceSpeakerIDsResp, error) {
	macID = strings.TrimSpace(macID)
	if macID == "" {
		return nil, ErrDeviceMacIDRequired
	}
	ids, err := s.repo.ListSpeakerIDsByMacID(ctx, macID)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []string{}
	}
	return &dto.DeviceSpeakerIDsResp{
		MacID:      macID,
		SpeakerIDs: ids,
	}, nil
}

// GetSpeakerNameByID returns the display name for a registered speaker.
func (s *DeviceAudioService) GetSpeakerNameByID(ctx context.Context, speakerID string) (*dto.DeviceSpeakerNameResp, error) {
	speakerID = strings.TrimSpace(speakerID)
	if speakerID == "" {
		return nil, ErrSpeakerIDRequired
	}
	row, err := s.speakerRepo.FindBySpeakerID(ctx, speakerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSpeakerNotFound
		}
		return nil, err
	}
	lines, err := util.ParseSpeakerPersonalLines(row.PersonalLines)
	if err != nil {
		lines = []string{}
	}
	return &dto.DeviceSpeakerNameResp{
		SpeakerID:     row.SpeakerID,
		SpeakerName:   row.SpeakerName,
		PersonalLines: lines,
	}, nil
}

// ListByMacID returns non-expired records for the device (no audio bytes in response).
func (s *DeviceAudioService) ListByMacID(ctx context.Context, macID string) ([]dto.DeviceAudioListItem, error) {
	macID = strings.TrimSpace(macID)
	if macID == "" {
		return nil, ErrDeviceMacIDRequired
	}
	rows, err := s.repo.ListVisibleByMacID(ctx, macID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DeviceAudioListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.DeviceAudioListItem{
			ID:         r.ID,
			MacID:      r.MacID,
			AudioLen:   r.AudioLen,
			Context:    r.Context,
			CreateTime: r.CreateTime,
			ExpireTime: r.ExpireTime,
		})
	}
	return out, nil
}

// PurgeDue deletes rows past delete_time.
func (s *DeviceAudioService) PurgeDue(ctx context.Context) (int64, error) {
	return s.repo.DeleteDue(ctx)
}

// RunCleanupLoop periodically purges due device audio rows until ctx is cancelled.
func (s *DeviceAudioService) RunCleanupLoop(ctx context.Context) {
	s.runPurge(ctx)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runPurge(ctx)
		}
	}
}

func (s *DeviceAudioService) runPurge(ctx context.Context) {
	n, err := s.PurgeDue(ctx)
	if err != nil {
		s.logger.Warn("device audio purge failed", zap.Error(err))
		return
	}
	if n > 0 {
		s.logger.Info("device audio purged", zap.Int64("rows", n))
	}
}

func isWAVPayload(data []byte, filename string) bool {
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return true
	}
	name := strings.ToLower(filename)
	return strings.HasSuffix(name, ".wav")
}
