package dto

type SysRouteItem struct {
	ID         string `json:"id"`
	ParentID   string `json:"parentId"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Title      string `json:"title"`
	I18nKey    string `json:"i18nKey"`
	Icon       string `json:"icon"`
	OrderNo    int    `json:"orderNo"`
	HideInMenu int    `json:"hideInMenu"`
	KeepAlive  int    `json:"keepAlive"`
	Status     int    `json:"status"`
}

type SysRouteCreateReq struct {
	ParentID   string `json:"parentId"`
	Name       string `json:"name" binding:"required"`
	Path       string `json:"path" binding:"required"`
	Component  string `json:"component" binding:"required"`
	Title      string `json:"title"`
	I18nKey    string `json:"i18nKey"`
	Icon       string `json:"icon"`
	OrderNo    int    `json:"orderNo"`
	HideInMenu int    `json:"hideInMenu"`
	KeepAlive  int    `json:"keepAlive"`
	Status     *int   `json:"status"`
}

type SysRouteUpdateReq struct {
	ID         string `json:"id" binding:"required"`
	ParentID   string `json:"parentId"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Title      string `json:"title"`
	I18nKey    string `json:"i18nKey"`
	Icon       string `json:"icon"`
	OrderNo    int    `json:"orderNo"`
	HideInMenu int    `json:"hideInMenu"`
	KeepAlive  int    `json:"keepAlive"`
	Status     int    `json:"status"`
}

type SysRouteDeleteReq struct {
	ID string `json:"id" binding:"required"`
}

type RoleSetRoutesReq struct {
	RoleID   string   `json:"roleId" binding:"required"`
	RouteIDs []string `json:"routeIds"`
}

type UserDeleteReq struct {
	ID string `json:"id" binding:"required"`
}

type RoleDeleteReq struct {
	ID string `json:"id" binding:"required"`
}
