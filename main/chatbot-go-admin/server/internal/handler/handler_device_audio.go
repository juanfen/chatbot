package handler

import (
	"errors"
	"io"
	"net/http"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) deviceAudioUpload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid multipart form")
		return
	}

	macID := c.PostForm("mac_id")
	audioLen := c.PostForm("audio_len")
	textContext := c.PostForm("context")

	file, err := c.FormFile("audio_blob")
	if err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "audio_blob is required")
		return
	}

	f, err := file.Open()
	if err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "open audio_blob failed")
		return
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "read audio_blob failed")
		return
	}

	resp, err := h.deviceAudioSvc.Upload(c.Request.Context(), macID, audioLen, textContext, data, file.Filename)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDeviceMacIDRequired),
			errors.Is(err, service.ErrDeviceAudioLenRequired),
			errors.Is(err, service.ErrDeviceAudioEmpty),
			errors.Is(err, service.ErrDeviceAudioTooLarge),
			errors.Is(err, service.ErrDeviceAudioNotWAV):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}

	common.OK(c, resp)
}

func (h *Handlers) deviceAudioList(c *gin.Context) {
	macID := c.Query("mac_id")
	list, err := h.deviceAudioSvc.ListByMacID(c.Request.Context(), macID)
	if err != nil {
		if errors.Is(err, service.ErrDeviceMacIDRequired) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct {
		List []dto.DeviceAudioListItem `json:"list"`
	}{List: list})
}

func (h *Handlers) deviceSpeakerIDsByMac(c *gin.Context) {
	macID := c.Query("mac_id")
	data, err := h.deviceAudioSvc.ListSpeakerIDsByMacID(c.Request.Context(), macID)
	if err != nil {
		if errors.Is(err, service.ErrDeviceMacIDRequired) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, data)
}

func (h *Handlers) deviceSpeakerNameByID(c *gin.Context) {
	speakerID := c.Query("speaker_id")
	data, err := h.deviceAudioSvc.GetSpeakerNameByID(c.Request.Context(), speakerID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSpeakerIDRequired):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrSpeakerNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, err.Error())
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, data)
}
