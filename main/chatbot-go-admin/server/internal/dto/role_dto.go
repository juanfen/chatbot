package dto

type RoleItem struct {
	ID       string   `json:"id"`
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Status   int      `json:"status"`
	RouteIDs []string `json:"routeIds"`
}

type RoleListResp struct {
	List []RoleItem `json:"list"`
}

type RoleCreateReq struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type RoleUpdateReq struct {
	ID     string `json:"id" binding:"required"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

