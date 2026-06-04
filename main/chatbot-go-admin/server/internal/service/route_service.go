package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RouteService struct {
	routeRepo      *repository.RouteRepository
	roleRouteRepo  *repository.RoleRouteRepository
	userRoleRepo   *repository.UserRoleRepository
	roleRepo       *repository.RoleRepository
}

func NewRouteService(
	routeRepo *repository.RouteRepository,
	roleRouteRepo *repository.RoleRouteRepository,
	userRoleRepo *repository.UserRoleRepository,
	roleRepo *repository.RoleRepository,
) *RouteService {
	return &RouteService{
		routeRepo:     routeRepo,
		roleRouteRepo: roleRouteRepo,
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
	}
}

// RouteTree returns admin management tree (legacy dto.RouteNode shape).
func (s *RouteService) RouteTree(ctx context.Context) (*dto.RouteTreeResp, error) {
	rows, err := s.routeRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.RouteTreeResp{Tree: buildRouteNodes(rows)}, nil
}

func (s *RouteService) MenuRoutesForUser(ctx context.Context, userID string, jwtIsAdmin bool) ([]dto.SAMenuRoute, string, error) {
	all, err := s.routeRepo.ListEnabled(ctx)
	if err != nil {
		return nil, "", err
	}
	byID := make(map[string]model.SysRoute, len(all))
	for _, r := range all {
		byID[r.ID] = r
	}

	allowed := make(map[string]struct{})
	if jwtIsAdmin {
		for _, r := range all {
			allowed[r.ID] = struct{}{}
		}
	} else {
		roleIDs, err := s.userRoleRepo.ListRoleIDsByUserID(ctx, userID)
		if err != nil {
			return nil, "", err
		}
		roles, err := s.roleRepo.FindByIDs(ctx, roleIDs)
		if err != nil {
			return nil, "", err
		}
		roleAdmin := false
		for _, ro := range roles {
			if ro.Code == "admin" {
				roleAdmin = true
				break
			}
		}
		if roleAdmin {
			for _, r := range all {
				allowed[r.ID] = struct{}{}
			}
		} else {
			for _, rid := range roleIDs {
				ids, err := s.roleRouteRepo.ListRouteIDsByRoleID(ctx, rid)
				if err != nil {
					return nil, "", err
				}
				for _, x := range ids {
					allowed[x] = struct{}{}
				}
			}
		}
	}

	for _, r := range all {
		if r.Name == "home" {
			allowed[r.ID] = struct{}{}
		}
	}

	allowed = expandRouteAncestors(byID, allowed)
	filtered := filterRoutesByIDs(all, allowed)
	routes := buildSAMenuRoutes(filtered)
	return routes, "home", nil
}

func expandRouteAncestors(byID map[string]model.SysRoute, seeds map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for id := range seeds {
		cur := id
		for cur != "" {
			r, ok := byID[cur]
			if !ok {
				break
			}
			out[cur] = struct{}{}
			cur = r.ParentID
		}
	}
	return out
}

func filterRoutesByIDs(all []model.SysRoute, allowed map[string]struct{}) []model.SysRoute {
	out := make([]model.SysRoute, 0, len(all))
	for _, r := range all {
		if _, ok := allowed[r.ID]; ok {
			out = append(out, r)
		}
	}
	return out
}

func buildRouteNodes(rows []model.SysRoute) []dto.RouteNode {
	childrenOf := map[string][]model.SysRoute{}
	for _, r := range rows {
		pid := r.ParentID
		childrenOf[pid] = append(childrenOf[pid], r)
	}
	var walk func(pid string) []dto.RouteNode
	walk = func(pid string) []dto.RouteNode {
		ch := childrenOf[pid]
		sort.Slice(ch, func(i, j int) bool {
			if ch[i].OrderNo == ch[j].OrderNo {
				return ch[i].ID < ch[j].ID
			}
			return ch[i].OrderNo < ch[j].OrderNo
		})
		out := make([]dto.RouteNode, 0, len(ch))
		for _, r := range ch {
			out = append(out, dto.RouteNode{
				Path:      r.Path,
				Name:      r.Name,
				Component: r.Component,
				Meta:      dto.RouteMeta{Title: r.Title, Icon: r.Icon},
				Children:  walk(r.ID),
			})
		}
		return out
	}
	return walk("")
}

func buildSAMenuRoutes(rows []model.SysRoute) []dto.SAMenuRoute {
	childrenOf := map[string][]model.SysRoute{}
	for _, r := range rows {
		pid := r.ParentID
		childrenOf[pid] = append(childrenOf[pid], r)
	}
	var walk func(pid string) []dto.SAMenuRoute
	walk = func(pid string) []dto.SAMenuRoute {
		ch := childrenOf[pid]
		sort.Slice(ch, func(i, j int) bool {
			if ch[i].OrderNo == ch[j].OrderNo {
				return ch[i].ID < ch[j].ID
			}
			return ch[i].OrderNo < ch[j].OrderNo
		})
		out := make([]dto.SAMenuRoute, 0, len(ch))
		for _, r := range ch {
			out = append(out, dto.SAMenuRoute{
				ID:        r.Name,
				Name:      r.Name,
				Path:      r.Path,
				Component: r.Component,
				Meta: dto.SAElegantRouteMeta{
					Title:      r.Title,
					I18nKey:    r.I18nKey,
					Icon:       r.Icon,
					Order:      r.OrderNo,
					HideInMenu: r.HideInMenu == 1,
					KeepAlive:  r.KeepAlive == 1,
				},
				Children: walk(r.ID),
			})
		}
		return out
	}
	return walk("")
}

// ListFlat returns all routes for CRUD table.
func (s *RouteService) ListFlat(ctx context.Context) ([]dto.SysRouteItem, error) {
	rows, err := s.routeRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SysRouteItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, sysRouteToItem(r))
	}
	return out, nil
}

func sysRouteToItem(r model.SysRoute) dto.SysRouteItem {
	return dto.SysRouteItem{
		ID:         r.ID,
		ParentID:   r.ParentID,
		Name:       r.Name,
		Path:       r.Path,
		Component:  r.Component,
		Title:      r.Title,
		I18nKey:    r.I18nKey,
		Icon:       r.Icon,
		OrderNo:    r.OrderNo,
		HideInMenu: r.HideInMenu,
		KeepAlive:  r.KeepAlive,
		Status:     r.Status,
	}
}

func (s *RouteService) CreateRoute(ctx context.Context, req dto.SysRouteCreateReq) error {
	if _, err := s.routeRepo.FindByName(ctx, req.Name); err == nil {
		return ErrRouteNameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if req.ParentID != "" {
		if _, err := s.routeRepo.FindByID(ctx, req.ParentID); err != nil {
			return err
		}
	}
	now := time.Now()
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	row := &model.SysRoute{
		ID:          uuid.NewString(),
		ParentID:    req.ParentID,
		Name:        req.Name,
		Path:        req.Path,
		Component:   req.Component,
		Title:       req.Title,
		I18nKey:     req.I18nKey,
		Icon:        req.Icon,
		OrderNo:     req.OrderNo,
		HideInMenu:  req.HideInMenu,
		KeepAlive:   req.KeepAlive,
		Status:      status,
		CreateTime:  now,
		UpdateTime:  now,
		Deleted:     0,
	}
	return s.routeRepo.Create(ctx, row)
}

func (s *RouteService) UpdateRoute(ctx context.Context, req dto.SysRouteUpdateReq) error {
	old, err := s.routeRepo.FindByID(ctx, req.ID)
	if err != nil {
		return err
	}
	pid := req.ParentID
	if pid != old.ParentID && pid != "" {
		if pid == req.ID {
			return errors.New("invalid parent")
		}
		if _, err := s.routeRepo.FindByID(ctx, pid); err != nil {
			return err
		}
	}
	path := req.Path
	if path == "" {
		path = old.Path
	}
	component := req.Component
	if component == "" {
		component = old.Component
	}
	updates := map[string]any{
		"parent_id":    pid,
		"path":         path,
		"component":    component,
		"title":        req.Title,
		"i18n_key":     req.I18nKey,
		"icon":         req.Icon,
		"order_no":     req.OrderNo,
		"hide_in_menu": req.HideInMenu,
		"keep_alive":   req.KeepAlive,
		"status":       req.Status,
		"update_time":  time.Now(),
	}
	return s.routeRepo.UpdateByID(ctx, req.ID, updates)
}

func (s *RouteService) DeleteRoute(ctx context.Context, id string) error {
	row, err := s.routeRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if row.Name == "home" {
		return ErrCannotDeleteHome
	}
	n, err := s.routeRepo.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrRouteHasChildren
	}
	if err := s.roleRouteRepo.SoftDeleteByRouteID(ctx, id); err != nil {
		return err
	}
	return s.routeRepo.SoftDelete(ctx, id)
}

// EnabledRouteNames for route existence check in Soybean.
func (s *RouteService) EnabledRouteNames(ctx context.Context) (map[string]struct{}, error) {
	rows, err := s.routeRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		m[r.Name] = struct{}{}
	}
	return m, nil
}
