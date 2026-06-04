package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"chatbot-go-admin/server/internal/common"
	"chatbot-go-admin/server/internal/config"
	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/middleware"
	"chatbot-go-admin/server/internal/repository"
	"chatbot-go-admin/server/internal/service"
	"chatbot-go-admin/server/internal/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handlers holds all HTTP handlers.
type Handlers struct {
	logger    *zap.Logger
	db        *gorm.DB
	jwtSecret string
	authSvc   *service.AuthService
	userSvc   *service.UserService
	roleSvc   *service.RoleService
	routeSvc  *service.RouteService
	bizSvc    *service.BizService
	rtCfgSvc  *service.RuntimeConfigService
	voiceSvc       *service.VoiceService
	deviceAudioSvc *service.DeviceAudioService

	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	userRoleRepo *repository.UserRoleRepository
}

// NewHandlers constructs handlers.
func NewHandlers(logger *zap.Logger, gormDB *gorm.DB, cfg config.Config) *Handlers {
	userRepo := repository.NewUserRepository(gormDB)
	roleRepo := repository.NewRoleRepository(gormDB)
	userRoleRepo := repository.NewUserRoleRepository(gormDB)
	routeRepo := repository.NewRouteRepository(gormDB)
	roleRouteRepo := repository.NewRoleRouteRepository(gormDB)
	pkgRepo := repository.NewBizPackageRepository(gormDB)
	quotaRepo := repository.NewBizUserQuotaRepository(gormDB)
	usageRepo := repository.NewBizUsageLogRepository(gormDB)
	annRepo := repository.NewBizAnnouncementRepository(gormDB)
	voiceRepo := repository.NewBizUserVoiceRepository(gormDB)
	prefRepo := repository.NewBizUserPreferenceRepository(gormDB)

	audioDir, err := cfg.ResolveAudioDir()
	if err != nil {
		logger.Warn("resolve audio dir failed, using configured path as-is", zap.Error(err))
		audioDir = cfg.AudioTargetPath
	}
	voiceSvc := service.NewVoiceService(logger, voiceRepo, prefRepo, audioDir, cfg.FFmpegBin, cfg.FFprobeBin, cfg.AudioDenoise)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	userSvc := service.NewUserService(userRepo, userRoleRepo, quotaRepo, pkgRepo)
	roleSvc := service.NewRoleService(roleRepo, roleRouteRepo, userRoleRepo)
	routeSvc := service.NewRouteService(routeRepo, roleRouteRepo, userRoleRepo, roleRepo)
	bizSvc := service.NewBizService(pkgRepo, quotaRepo, usageRepo, annRepo)
	rtCfgSvc := service.NewRuntimeConfigService(cfg.XiaoZhiCoreBaseURL)
	deviceAudioRepo := repository.NewBizDeviceAudioRepository(gormDB)
	speakerRepo := repository.NewBizUserSpeakerRepository(gormDB)
	deviceAudioSvc := service.NewDeviceAudioService(deviceAudioRepo, speakerRepo, service.NewVoiceprintClientFromConfig(cfg, logger), service.NewProfileImportClientFromConfig(cfg, logger), logger)
	return &Handlers{
		logger:    logger,
		db:        gormDB,
		jwtSecret: cfg.JWTSecret,
		authSvc:   authSvc,
		userSvc:   userSvc,
		roleSvc:   roleSvc,
		routeSvc:  routeSvc,
		bizSvc:    bizSvc,
		rtCfgSvc:  rtCfgSvc,
		voiceSvc:       voiceSvc,
		deviceAudioSvc: deviceAudioSvc,
		userRepo:       userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

// Register registers routes to gin engine.
func (h *Handlers) Register(r *gin.Engine) {
	// Primary API group
	h.registerV1(r.Group("/api/v1"))
	// Compatibility: allow calling without "/api/v1" prefix (for SoybeanAdmin web default "/auth/*", "/route/*")
	h.registerV1(r.Group(""))
}

func (h *Handlers) registerV1(g *gin.RouterGroup) {
	{
		g.GET("/health", h.health)

		// SoybeanAdmin-compatible auth endpoints
		g.POST("/auth/login", h.saLogin)

		// 设备上报 WAV 音频（无需登录）
		g.POST("/device/audio/upload", h.deviceAudioUpload)
		g.GET("/device/audio/list", h.deviceAudioList)
		g.GET("/device/audio/speaker-ids", h.deviceSpeakerIDsByMac)
		g.GET("/device/audio/speaker-name", h.deviceSpeakerNameByID)
	}

	protected := g.Group("")
	protected.Use(middleware.JWTAuth(h.jwtSecret))
	{
		// SoybeanAdmin-compatible auth endpoints
		protected.GET("/auth/getUserInfo", h.saGetUserInfo)
		protected.POST("/auth/refreshToken", h.saRefreshToken)

		// SoybeanAdmin-compatible route endpoints (dynamic route mode)
		protected.GET("/route/getUserRoutes", h.saGetUserRoutes)
		protected.GET("/route/isRouteExist", h.saIsRouteExist)

		// existing sys management endpoints
		protected.GET("/sys/user/list", h.userList)
		protected.POST("/sys/user/create", h.userCreate)
		protected.POST("/sys/user/update", h.userUpdate)

		protected.GET("/sys/role/list", h.roleList)
		protected.POST("/sys/role/create", h.roleCreate)
		protected.POST("/sys/role/update", h.roleUpdate)

		protected.GET("/sys/routes/tree", h.routesTree)
		protected.GET("/sys/route/list", h.routeListFlat)
		protected.POST("/sys/route/create", h.routeCreate)
		protected.POST("/sys/route/update", h.routeUpdate)
		protected.POST("/sys/route/delete", h.routeDelete)

		protected.POST("/sys/user/delete", h.userDelete)
		protected.POST("/sys/role/delete", h.roleDelete)
		protected.POST("/sys/role/setRoutes", h.roleSetRoutes)

		// 用户侧：额度、调用、公告、套餐目录、模拟购买、体验调用
		protected.GET("/biz/quota/summary", h.bizQuotaSummary)
		protected.GET("/biz/usage/trend", h.bizUsageTrend)
		protected.GET("/biz/usage/list", h.bizUsageList)
		protected.GET("/biz/announcements", h.bizAnnouncements)
		protected.GET("/biz/packages/catalog", h.bizPackageCatalog)
		protected.POST("/biz/purchase/simulate", h.bizPurchaseSimulate)
		protected.POST("/biz/usage/simulateCall", h.bizSimulateCall)

		// 管理：套餐与公告
		protected.GET("/sys/package/list", h.sysPackageList)
		protected.POST("/sys/package/create", h.sysPackageCreate)
		protected.POST("/sys/package/update", h.sysPackageUpdate)
		protected.POST("/sys/package/delete", h.sysPackageDelete)
		protected.GET("/sys/announcement/list", h.sysAnnouncementList)
		protected.POST("/sys/announcement/create", h.sysAnnouncementCreate)
		protected.POST("/sys/announcement/update", h.sysAnnouncementUpdate)
		protected.POST("/sys/announcement/delete", h.sysAnnouncementDelete)

		// 算法运行时配置：反向代理到 xiaozhi-go-core /xiaozhi/v1/runtime/config
		protected.GET("/sys/runtime-config", h.sysRuntimeConfigGet)
		protected.POST("/sys/runtime-config", h.sysRuntimeConfigPatch)

		// 移动端：标准文本、语音样本上传与激活
		protected.GET("/mobile/standard-text", h.mobileGetStandardText)
		protected.PUT("/mobile/standard-text", h.mobilePutStandardText)
		protected.GET("/mobile/voices", h.mobileListVoices)
		protected.POST("/mobile/voices/upload", h.mobileUploadVoice)
		protected.POST("/mobile/voices/:id/activate", h.mobileActivateVoice)
		protected.GET("/mobile/voices/:id/file", h.mobileVoiceFile)

		// 音色识别：设备音频管理与音色绑定
		protected.GET("/sys/device-audio/list", h.sysDeviceAudioList)
		protected.GET("/sys/device-audio/speakers", h.sysDeviceAudioSpeakers)
		protected.GET("/sys/device-audio/speakers/:speakerId/personal-lines", h.sysDeviceAudioSpeakerPersonalGet)
		protected.POST("/sys/device-audio/speakers/:speakerId/personal-lines", h.sysDeviceAudioSpeakerPersonalPut)
		protected.POST("/sys/device-audio/bind", h.sysDeviceAudioBind)
		protected.GET("/sys/device-audio/:id/file", h.sysDeviceAudioFile)
	}

	// constant routes should be accessible before login in dynamic route mode
	g.GET("/route/getConstantRoutes", h.saGetConstantRoutes)
}

func (h *Handlers) health(c *gin.Context) {
	common.OK(c, &struct {
		Status string `json:"status"`
	}{Status: "ok"})
}

func (h *Handlers) notImplemented(c *gin.Context) {
	common.Fail(c, http.StatusNotImplemented, common.CodeInternal, "not implemented")
}

// saLogin is compatible with SoybeanAdmin's request payload and response shape.
func (h *Handlers) saLogin(c *gin.Context) {
	var req dto.SALoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	username := req.UserName
	if username == "" {
		username = req.Username
	}
	if username == "" {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "userName is required")
		return
	}

	token, userID, isAdmin, err := h.authSvc.Login(c.Request.Context(), username, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, "invalid credentials")
			return
		case service.ErrUserDisabled:
			common.Fail(c, http.StatusForbidden, common.CodeForbidden, "user disabled")
			return
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
			return
		}
	}

	// For now, use a JWT as refresh token as well (longer TTL).
	refreshToken, err := util.SignToken(h.jwtSecret, userID, isAdmin, 30*24*time.Hour)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}

	common.OK(c, &dto.SALoginToken{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (h *Handlers) saRefreshToken(c *gin.Context) {
	var req dto.SARefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}

	claims, err := util.ParseToken(h.jwtSecret, req.RefreshToken)
	if err != nil {
		common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, "invalid refresh token")
		return
	}

	accessToken, err := util.SignToken(h.jwtSecret, claims.UserID, claims.IsAdmin, 7*24*time.Hour)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	refreshToken, err := util.SignToken(h.jwtSecret, claims.UserID, claims.IsAdmin, 30*24*time.Hour)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}

	common.OK(c, &dto.SALoginToken{Token: accessToken, RefreshToken: refreshToken})
}

func (h *Handlers) saGetUserInfo(c *gin.Context) {
	userIDAny, _ := c.Get(middleware.CtxKeyUserID)
	userID, _ := userIDAny.(string)
	if userID == "" {
		common.Fail(c, http.StatusUnauthorized, common.CodeUnauthorized, "unauthorized")
		return
	}

	u, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}

	roleIDs, err := h.userRoleRepo.ListRoleIDsByUserID(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	roles, err := h.roleRepo.FindByIDs(c.Request.Context(), roleIDs)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}

	roleCodes := make([]string, 0, len(roles)+1)
	for _, r := range roles {
		if r.Code != "" {
			roleCodes = append(roleCodes, r.Code)
		}
	}
	if u.IsAdmin == 1 {
		// keep compatible with frontend env VITE_STATIC_SUPER_ROLE=super
		roleCodes = append(roleCodes, "admin", "super")
	}

	userName := u.Nickname
	if userName == "" {
		userName = u.Username
	}

	common.OK(c, &dto.SAUserInfoResp{
		UserID:   u.ID,
		UserName: userName,
		Roles:    roleCodes,
		Buttons:  []string{},
	})
}

func (h *Handlers) saGetConstantRoutes(c *gin.Context) {
	// Minimal constant routes for dynamic route mode
	routes := []dto.SAMenuRoute{
		{
			ID:        "login",
			Name:      "login",
			Path:      "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			Component: "layout.blank$view.login",
			Meta: dto.SAElegantRouteMeta{
				Title:      "login",
				I18nKey:    "route.login",
				Constant:   true,
				HideInMenu: true,
			},
		},
		{
			ID:        "403",
			Name:      "403",
			Path:      "/403",
			Component: "layout.blank$view.403",
			Meta: dto.SAElegantRouteMeta{
				Title:      "403",
				I18nKey:    "route.403",
				Constant:   true,
				HideInMenu: true,
			},
		},
		{
			ID:        "404",
			Name:      "404",
			Path:      "/404",
			Component: "layout.blank$view.404",
			Meta: dto.SAElegantRouteMeta{
				Title:      "404",
				I18nKey:    "route.404",
				Constant:   true,
				HideInMenu: true,
			},
		},
		{
			ID:        "500",
			Name:      "500",
			Path:      "/500",
			Component: "layout.blank$view.500",
			Meta: dto.SAElegantRouteMeta{
				Title:      "500",
				I18nKey:    "route.500",
				Constant:   true,
				HideInMenu: true,
			},
		},
		{
			ID:        "iframe-page",
			Name:      "iframe-page",
			Path:      "/iframe-page/:url",
			Component: "layout.base$view.iframe-page",
			Props:     true,
			Meta: dto.SAElegantRouteMeta{
				Title:      "iframe-page",
				I18nKey:    "route.iframe-page",
				Constant:   true,
				HideInMenu: true,
				KeepAlive:  true,
			},
		},
	}

	common.OK(c, &routes)
}

func (h *Handlers) saGetUserRoutes(c *gin.Context) {
	userIDAny, _ := c.Get(middleware.CtxKeyUserID)
	isAdminAny, _ := c.Get(middleware.CtxKeyIsAdmin)
	userID, _ := userIDAny.(string)
	isAdmin, _ := isAdminAny.(bool)
	routes, home, err := h.routeSvc.MenuRoutesForUser(c.Request.Context(), userID, isAdmin)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &dto.SAUserRouteResp{Routes: routes, Home: home})
}

func (h *Handlers) saIsRouteExist(c *gin.Context) {
	routeName := c.Query("routeName")
	if routeName == "" {
		common.OK(c, ptrBool(false))
		return
	}
	switch routeName {
	case "login", "403", "404", "500", "iframe-page":
		common.OK(c, ptrBool(true))
		return
	}
	names, err := h.routeSvc.EnabledRouteNames(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	_, ok := names[routeName]
	common.OK(c, ptrBool(ok))
}

func ptrBool(v bool) *bool { return &v }

func (h *Handlers) userList(c *gin.Context) {
	var req dto.UserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	resp, err := h.userSvc.List(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, resp)
}

func (h *Handlers) userCreate(c *gin.Context) {
	var req dto.UserCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.userSvc.Create(c.Request.Context(), req); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "user already exists")
			return
		}
		if errors.Is(err, service.ErrPackageNotFound) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "package not found")
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) userUpdate(c *gin.Context) {
	var req dto.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.userSvc.Update(c.Request.Context(), req); err != nil {
		if errors.Is(err, service.ErrPackageNotFound) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "package not found")
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) roleList(c *gin.Context) {
	resp, err := h.roleSvc.List(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, resp)
}

func (h *Handlers) roleCreate(c *gin.Context) {
	var req dto.RoleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.roleSvc.Create(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) roleUpdate(c *gin.Context) {
	var req dto.RoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.roleSvc.Update(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) routesTree(c *gin.Context) {
	resp, err := h.routeSvc.RouteTree(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, resp)
}

func (h *Handlers) routeListFlat(c *gin.Context) {
	list, err := h.routeSvc.ListFlat(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct {
		List []dto.SysRouteItem `json:"list"`
	}{List: list})
}

func (h *Handlers) routeCreate(c *gin.Context) {
	var req dto.SysRouteCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.routeSvc.CreateRoute(c.Request.Context(), req); err != nil {
		if errors.Is(err, service.ErrRouteNameExists) {
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) routeUpdate(c *gin.Context) {
	var req dto.SysRouteUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.routeSvc.UpdateRoute(c.Request.Context(), req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
			return
		}
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) routeDelete(c *gin.Context) {
	var req dto.SysRouteDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.routeSvc.DeleteRoute(c.Request.Context(), req.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotDeleteHome):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrRouteHasChildren):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) userDelete(c *gin.Context) {
	var req dto.UserDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	op, _ := c.Get(middleware.CtxKeyUserID)
	opID, _ := op.(string)
	if err := h.userSvc.Delete(c.Request.Context(), opID, req.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotDeleteSelf):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrLastAdminUser):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) roleDelete(c *gin.Context) {
	var req dto.RoleDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.roleSvc.Delete(c.Request.Context(), req.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrReservedRole):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, service.ErrRoleInUse):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, err.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) roleSetRoutes(c *gin.Context) {
	var req dto.RoleSetRoutesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.roleSvc.SetRoutes(c.Request.Context(), req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.Fail(c, http.StatusNotFound, common.CodeNotFound, "not found")
			return
		}
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) ctxUser(c *gin.Context) (userID string, isAdmin bool) {
	uid, _ := c.Get(middleware.CtxKeyUserID)
	userID, _ = uid.(string)
	ia, _ := c.Get(middleware.CtxKeyIsAdmin)
	isAdmin, _ = ia.(bool)
	return userID, isAdmin
}

func (h *Handlers) requireAdmin(c *gin.Context) bool {
	_, isAdmin := h.ctxUser(c)
	return isAdmin
}

func (h *Handlers) bizQuotaSummary(c *gin.Context) {
	userID, isAdmin := h.ctxUser(c)
	data, err := h.bizSvc.QuotaSummary(c.Request.Context(), userID, isAdmin)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, data)
}

func (h *Handlers) bizUsageTrend(c *gin.Context) {
	userID, isAdmin := h.ctxUser(c)
	days := 0
	if v := c.Query("days"); v != "" {
		if n, err := parseInt(v); err == nil {
			days = n
		}
	}
	data, err := h.bizSvc.UsageTrend(c.Request.Context(), userID, isAdmin, days)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &data)
}

func (h *Handlers) bizUsageList(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	page, pageSize := 1, 20
	if v := c.Query("page"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("pageSize"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			pageSize = n
		}
	}
	data, err := h.bizSvc.UsageList(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, data)
}

func (h *Handlers) bizAnnouncements(c *gin.Context) {
	data, err := h.bizSvc.Announcements(c.Request.Context(), 20)
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &data)
}

func (h *Handlers) bizPackageCatalog(c *gin.Context) {
	data, err := h.bizSvc.PackageCatalog(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &data)
}

func (h *Handlers) bizPurchaseSimulate(c *gin.Context) {
	userID, _ := h.ctxUser(c)
	var req dto.BizPurchaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	data, err := h.bizSvc.SimulatedPurchase(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPackageNotFound):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "package not found")
		case errors.Is(err, service.ErrPackageDisabled):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "package disabled")
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, data)
}

func (h *Handlers) bizSimulateCall(c *gin.Context) {
	userID, isAdmin := h.ctxUser(c)
	if err := h.bizSvc.SimulateExperienceCall(c.Request.Context(), userID, isAdmin); err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientQuota):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "insufficient quota")
		case errors.Is(err, service.ErrUserQuotaNotFound):
			common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "quota not configured")
		default:
			common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		}
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysPackageList(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	data, err := h.bizSvc.AdminPackageList(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &data)
}

func (h *Handlers) sysPackageCreate(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req dto.SysPackageCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminPackageCreate(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysPackageUpdate(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req dto.SysPackageUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminPackageUpdate(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysPackageDelete(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminPackageDelete(c.Request.Context(), req.ID); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysAnnouncementList(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	data, err := h.bizSvc.AdminAnnouncementList(c.Request.Context())
	if err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &data)
}

func (h *Handlers) sysAnnouncementCreate(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req dto.SysAnnouncementCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminAnnouncementCreate(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysAnnouncementUpdate(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req dto.SysAnnouncementUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminAnnouncementUpdate(c.Request.Context(), req); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func (h *Handlers) sysAnnouncementDelete(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var req struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid request")
		return
	}
	if err := h.bizSvc.AdminAnnouncementDelete(c.Request.Context(), req.ID); err != nil {
		common.Fail(c, http.StatusInternalServerError, common.CodeInternal, "internal error")
		return
	}
	common.OK(c, &struct{}{})
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// sysRuntimeConfigGet 透传上游当前配置（仅 8 个允许键）。
func (h *Handlers) sysRuntimeConfigGet(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	data, err := h.rtCfgSvc.Get(c.Request.Context())
	if err != nil {
		h.logger.Warn("runtime config get failed", zap.Error(err))
		common.Fail(c, http.StatusBadGateway, common.CodeInternal, "upstream unavailable: "+err.Error())
		return
	}
	common.OK(c, &data)
}

// sysRuntimeConfigPatch 把补丁透传给上游（深度合并）。
func (h *Handlers) sysRuntimeConfigPatch(c *gin.Context) {
	if !h.requireAdmin(c) {
		common.Fail(c, http.StatusForbidden, common.CodeForbidden, "forbidden")
		return
	}
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		common.Fail(c, http.StatusBadRequest, common.CodeBadRequest, "invalid json: "+err.Error())
		return
	}
	if err := h.rtCfgSvc.Patch(c.Request.Context(), patch); err != nil {
		h.logger.Warn("runtime config patch failed", zap.Error(err))
		common.Fail(c, http.StatusBadGateway, common.CodeInternal, "upstream rejected: "+err.Error())
		return
	}
	common.OK(c, &struct{}{})
}

