package handler

import (
	"errors"
	"net/http"
	"strconv"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/middleware"
	"chatbot-go-admin/server/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) sysDeviceAudioList(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxKeyUserID)
	uid, _ := userID.(string)

	page, pageSize := 1, 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("pageSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}

	data, err := h.deviceAudioSvc.ListManageableDeviceAudio(c.Request.Context(), uid, page, pageSize)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, data)
}

func (h *Handlers) sysDeviceAudioSpeakers(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxKeyUserID)
	uid, _ := userID.(string)

	list, err := h.deviceAudioSvc.ListUserSpeakers(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, err.Error())
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct {
		List []dto.UserSpeakerItem `json:"list"`
	}{List: list})
}

func (h *Handlers) sysDeviceAudioBind(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxKeyUserID)
	uid, _ := userID.(string)

	var req dto.DeviceAudioBindReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}

	resp, err := h.deviceAudioSvc.BindDeviceAudioSpeaker(c.Request.Context(), uid, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, err.Error())
		case errors.Is(err, service.ErrDeviceAudioIDRequired),
			errors.Is(err, service.ErrSpeakerNameOrIDRequired),
			errors.Is(err, service.ErrDeviceAudioAlreadyBound),
			errors.Is(err, service.ErrDeviceAudioExpired):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrDeviceAudioNotFound),
			errors.Is(err, service.ErrSpeakerNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, err.Error())
		case errors.Is(err, service.ErrVoiceprintNotConfigured):
			common.Fail(c, http.StatusBadGateway, common.CodeInternal, err.Error())
		case errors.Is(err, service.ErrVoiceprintReportFailed):
			common.Fail(c, http.StatusBadGateway, common.CodeInternal, err.Error())
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, resp)
}

func (h *Handlers) sysDeviceAudioSpeakerPersonalGet(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxKeyUserID)
	uid, _ := userID.(string)
	speakerID := c.Param("speakerId")

	data, err := h.deviceAudioSvc.GetSpeakerPersonalLines(c.Request.Context(), uid, speakerID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, err.Error())
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

func (h *Handlers) sysDeviceAudioSpeakerPersonalPut(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxKeyUserID)
	uid, _ := userID.(string)
	speakerID := c.Param("speakerId")

	var req dto.SpeakerPersonalLinesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}

	data, err := h.deviceAudioSvc.SetSpeakerPersonalLines(c.Request.Context(), uid, speakerID, req.Lines)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, err.Error())
		case errors.Is(err, service.ErrSpeakerIDRequired),
			errors.Is(err, service.ErrSpeakerPersonalInvalid):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrSpeakerNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, err.Error())
		case errors.Is(err, service.ErrSpeakerMacNotFound):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrProfileImportFailed):
			common.Fail(c, http.StatusBadGateway, common.CodeInternal, err.Error())
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, data)
}

func (h *Handlers) sysDeviceAudioFile(c *gin.Context) {
	id := c.Param("id")
	data, err := h.deviceAudioSvc.GetDeviceAudioBlob(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDeviceAudioIDRequired),
			errors.Is(err, service.ErrDeviceAudioExpired):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrDeviceAudioNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, err.Error())
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	c.Header("Content-Type", "audio/wav")
	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, "audio/wav", data)
}
