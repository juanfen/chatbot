package dto

type UserListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
}

type UserItem struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	Status       int      `json:"status"`
	IsAdmin      int      `json:"isAdmin"`
	RoleIDs      []string `json:"roleIds"`
	TotalQuota   int64    `json:"totalQuota"`
	UsedQuota    int64    `json:"usedQuota"`
	PackageTitle string   `json:"packageTitle"`
}

type UserListResp struct {
	Total int64      `json:"total"`
	List  []UserItem `json:"list"`
}

type UserCreateReq struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required"`
	Nickname string   `json:"nickname"`
	IsAdmin  int      `json:"isAdmin"`
	RoleIDs  []string `json:"roleIds"`
	// TrialQuota: 体验调用次数（与 DefaultPackageID 二选一或叠加时优先套餐）
	TrialQuota *int64 `json:"trialQuota"`
	// DefaultPackageID: 创建后按套餐额度初始化（biz_package.id）
	DefaultPackageID string `json:"defaultPackageId"`
}

type UserUpdateReq struct {
	ID       string   `json:"id" binding:"required"`
	Nickname string   `json:"nickname"`
	Password string   `json:"password"`
	Status   int      `json:"status"`
	IsAdmin  int      `json:"isAdmin"`
	RoleIDs  []string `json:"roleIds"`
	TrialQuota       *int64 `json:"trialQuota"`
	DefaultPackageID string `json:"defaultPackageId"`
}

