package dto

type RouteMeta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type RouteNode struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Component string      `json:"component"`
	Redirect  string      `json:"redirect,omitempty"`
	Meta      RouteMeta   `json:"meta"`
	Children  []RouteNode `json:"children,omitempty"`
}

type RouteTreeResp struct {
	Tree []RouteNode `json:"tree"`
}

// SoybeanAdmin compatibility DTOs (used by web/src/service/api/route.ts)
// TypeScript side expects ElegantConstRoute-like objects.
type SAElegantRouteMeta struct {
	Title      string   `json:"title,omitempty"`
	I18nKey    string   `json:"i18nKey,omitempty"`
	Icon       string   `json:"icon,omitempty"`
	Order      int      `json:"order,omitempty"`
	Constant   bool     `json:"constant,omitempty"`
	HideInMenu bool     `json:"hideInMenu,omitempty"`
	KeepAlive  bool     `json:"keepAlive,omitempty"`
	Href       string   `json:"href,omitempty"`
	ActiveMenu string   `json:"activeMenu,omitempty"`
	Roles      []string `json:"roles,omitempty"`
}

type SAMenuRoute struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Path      string            `json:"path"`
	Component string            `json:"component,omitempty"`
	Redirect  string            `json:"redirect,omitempty"`
	Props     any               `json:"props,omitempty"`
	Meta      SAElegantRouteMeta `json:"meta,omitempty"`
	Children  []SAMenuRoute      `json:"children,omitempty"`
}

type SAUserRouteResp struct {
	Routes []SAMenuRoute `json:"routes"`
	Home   string        `json:"home"`
}

