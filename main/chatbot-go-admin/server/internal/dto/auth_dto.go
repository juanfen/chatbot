package dto

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token string `json:"token"`
}

// SoybeanAdmin compatibility DTOs (used by web/src/service/api/auth.ts)
// - login: { userName, password }
// - response: { token, refreshToken }
type SALoginReq struct {
	UserName string `json:"userName"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type SALoginToken struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type SARefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type SAUserInfoResp struct {
	UserID   string   `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
	Buttons  []string `json:"buttons"`
}

