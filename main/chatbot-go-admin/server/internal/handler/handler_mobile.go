package handler

import (
	"errors"
	"net/http"
	"os"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handlers) mobileGetStandardText(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	text, err := h.voiceSvc.GetStandardText(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &dto.MobileStandardTextResp{StandardText: text})
}

func (h *Handlers) mobilePutStandardText(c *gin.Context) {
	var req dto.MobileStandardTextReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	userID, _ := h.ctxUser(c)
	if err := h.voiceSvc.SetStandardText(c.Request.Context(), userID, req.StandardText); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) mobileListVoices(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	list, err := h.voiceSvc.ListVoices(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &list)
}

func (h *Handlers) mobileUploadVoice(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid multipart form")
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "file is required")
		return
	}

	tmpFile, err := os.CreateTemp("", "voice-upload-*")
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "temp file failed")
		return
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := c.SaveUploadedFile(fh, tmpPath); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "save upload failed")
		return
	}

	item, err := h.voiceSvc.UploadConvertSave(c.Request.Context(), userID, tmpPath, fh.Filename)
	if err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		return
	}
	common.OK(c, item)
}

func (h *Handlers) mobileActivateVoice(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	id := c.Param("id")
	if id == "" {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "id is required")
		return
	}
	if err := h.voiceSvc.Activate(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "voice not found")
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) mobileVoiceFile(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	id := c.Param("id")
	if id == "" {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "id is required")
		return
	}
	abs, err := h.voiceSvc.AbsPathForUserVoice(c.Request.Context(), userID, id)
	if err != nil {
		common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
		return
	}
	c.Header("Content-Type", "audio/wav")
	c.File(abs)
}
