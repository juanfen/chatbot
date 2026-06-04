package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DefaultStandardText 为未配置时的默认朗读稿（录音页上半屏展示）。
const DefaultStandardText = `欢迎使用智能语音服务。请您用自然、平稳的语速朗读以下句子：
春天来了，花园里的花儿竞相开放。微风拂过树梢，鸟儿在枝头轻声歌唱。
科技的发展让生活更加便利，我们也期待用声音连接彼此，传递温暖与关怀。`

var (
	wavFileNameRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.wav$`)
)

// speechDenoiseFilter 面向人声样本的 ffmpeg 滤镜链（需 libavfilter 含 afftdn）。
// highpass/lowpass 去底噪与高频嘶声；afftdn 频域降噪；agate 压掉静音段底噪；loudnorm 统一响度。
const speechDenoiseFilter = "highpass=f=100,lowpass=f=7500,afftdn=nr=15:nf=-28:tn=1," +
	"agate=threshold=0.008:ratio=3:attack=150:release=900," +
	"loudnorm=I=-18:TP=-1.5:LRA=11"

// VoiceService handles voice uploads, ffmpeg conversion, and preferences.
type VoiceService struct {
	logger      *zap.Logger
	voices      *repository.BizUserVoiceRepository
	prefs       *repository.BizUserPreferenceRepository
	audioDir    string
	ffmpeg      string
	ffprobe     string
	audioDenoise bool
}

func NewVoiceService(
	logger *zap.Logger,
	voices *repository.BizUserVoiceRepository,
	prefs *repository.BizUserPreferenceRepository,
	audioDir, ffmpeg, ffprobe string,
	audioDenoise bool,
) *VoiceService {
	return &VoiceService{
		logger:       logger,
		voices:       voices,
		prefs:        prefs,
		audioDir:     audioDir,
		ffmpeg:       ffmpeg,
		ffprobe:      ffprobe,
		audioDenoise: audioDenoise,
	}
}

func (s *VoiceService) ensureAudioDir() error {
	return os.MkdirAll(s.audioDir, 0o755)
}

func (s *VoiceService) GetStandardText(ctx context.Context, userID string) (string, error) {
	row, err := s.prefs.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultStandardText, nil
		}
		return "", err
	}
	t := strings.TrimSpace(row.StandardText)
	if t == "" {
		return DefaultStandardText, nil
	}
	return row.StandardText, nil
}

func (s *VoiceService) SetStandardText(ctx context.Context, userID, text string) error {
	if len([]rune(text)) > 8000 {
		return fmt.Errorf("standard text too long (max 8000 characters)")
	}
	return s.prefs.UpsertStandardText(ctx, userID, text)
}

func (s *VoiceService) ListVoices(ctx context.Context, userID string) ([]dto.MobileVoiceItem, error) {
	rows, err := s.voices.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MobileVoiceItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.MobileVoiceItem{
			ID:              r.ID,
			FileName:        r.FileName,
			UploadTime:      r.UploadTime,
			DurationSeconds: r.DurationSeconds,
			IsActive:        r.IsActive == 1,
		})
	}
	return out, nil
}

func (s *VoiceService) Activate(ctx context.Context, userID, voiceID string) error {
	return s.voices.ActivateOne(ctx, userID, voiceID)
}

func (s *VoiceService) ResolveStoredPath(fileName string) (string, error) {
	base := filepath.Base(strings.TrimSpace(fileName))
	if !wavFileNameRe.MatchString(base) {
		return "", fmt.Errorf("invalid file name")
	}
	return filepath.Join(s.audioDir, base), nil
}

func (s *VoiceService) AbsPathForUserVoice(ctx context.Context, userID, voiceID string) (string, error) {
	row, err := s.voices.FindByIDAndUser(ctx, voiceID, userID)
	if err != nil {
		return "", err
	}
	return s.ResolveStoredPath(row.FileName)
}

// UploadConvertSave saves an uploaded temp file as 24kHz mono WAV under audioDir.
func (s *VoiceService) UploadConvertSave(ctx context.Context, userID, tempInputPath, origName string) (*dto.MobileVoiceItem, error) {
	ext := strings.ToLower(filepath.Ext(origName))
	if ext == "" {
		ext = ".bin"
	}
	if !allowedAudioExt[ext] {
		return nil, fmt.Errorf("unsupported audio type %s (allowed: mp3 wav m4a aac webm ogg flac opus)", ext)
	}
	if err := s.ensureAudioDir(); err != nil {
		return nil, err
	}

	id := uuid.NewString()
	outName := id + ".wav"
	outAbs := filepath.Join(s.audioDir, outName)

	if err := s.convertToVoiceWAV(ctx, tempInputPath, outAbs); err != nil {
		_ = os.Remove(outAbs)
		return nil, err
	}

	dur, err := probeDuration(ctx, s.ffprobe, outAbs)
	if err != nil {
		_ = os.Remove(outAbs)
		return nil, err
	}

	now := time.Now()
	row := &model.BizUserVoice{
		ID:              id,
		UserID:          userID,
		FileName:        outName,
		UploadTime:      now,
		DurationSeconds: dur,
		IsActive:        0,
		CreateTime:      now,
		UpdateTime:      now,
		Deleted:         0,
	}
	if err := s.voices.Create(ctx, row); err != nil {
		_ = os.Remove(outAbs)
		return nil, err
	}

	return &dto.MobileVoiceItem{
		ID:              id,
		FileName:        outName,
		UploadTime:      now,
		DurationSeconds: dur,
		IsActive:        false,
	}, nil
}

func (s *VoiceService) convertToVoiceWAV(ctx context.Context, inputPath, outAbs string) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", inputPath}
	if s.audioDenoise {
		args = append(args, "-af", speechDenoiseFilter)
	}
	args = append(args, "-ar", "24000", "-ac", "1", "-c:a", "pcm_s16le", outAbs)

	out, err := exec.CommandContext(ctx, s.ffmpeg, args...).CombinedOutput()
	if err == nil {
		return nil
	}

	// 降噪滤镜在部分 ffmpeg 构建上不可用；回退为仅转码，避免上传完全失败
	if s.audioDenoise {
		s.logger.Warn("ffmpeg denoise convert failed, retry without filters",
			zap.Error(err), zap.String("stderr", string(out)))
		fallback := []string{
			"-hide_banner", "-loglevel", "error", "-y", "-i", inputPath,
			"-ar", "24000", "-ac", "1", "-c:a", "pcm_s16le", outAbs,
		}
		out2, err2 := exec.CommandContext(ctx, s.ffmpeg, fallback...).CombinedOutput()
		if err2 == nil {
			return nil
		}
		s.logger.Warn("ffmpeg fallback convert failed", zap.Error(err2), zap.String("stderr", string(out2)))
		return fmt.Errorf("audio convert failed (ensure ffmpeg is installed): %w", err2)
	}

	s.logger.Warn("ffmpeg convert failed", zap.Error(err), zap.String("stderr", string(out)))
	return fmt.Errorf("audio convert failed (ensure ffmpeg is installed): %w", err)
}

var allowedAudioExt = map[string]bool{
	".mp3": true, ".wav": true, ".m4a": true, ".aac": true, ".webm": true,
	".ogg": true, ".flac": true, ".opus": true,
}

func probeDuration(ctx context.Context, ffprobe, path string) (float64, error) {
	cmd := exec.CommandContext(ctx, ffprobe, "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	s := strings.TrimSpace(string(out))
	if s == "" || s == "N/A" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return f, nil
}
