package service

import (
	"context"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"github.com/google/uuid"
)

type RoleService struct {
	roleRepo      *repository.RoleRepository
	roleRouteRepo *repository.RoleRouteRepository
	userRoleRepo  *repository.UserRoleRepository
}

func NewRoleService(
	roleRepo *repository.RoleRepository,
	roleRouteRepo *repository.RoleRouteRepository,
	userRoleRepo *repository.UserRoleRepository,
) *RoleService {
	return &RoleService{
		roleRepo:      roleRepo,
		roleRouteRepo: roleRouteRepo,
		userRoleRepo:  userRoleRepo,
	}
}

func (s *RoleService) List(ctx context.Context) (*dto.RoleListResp, error) {
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(roles))
	for i, r := range roles {
		ids[i] = r.ID
	}
	routeMap, err := s.roleRouteRepo.ListRouteIDsByRoleIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RoleItem, 0, len(roles))
	for _, r := range roles {
		out = append(out, dto.RoleItem{
			ID:       r.ID,
			Code:     r.Code,
			Name:     r.Name,
			Status:   r.Status,
			RouteIDs: routeMap[r.ID],
		})
	}
	return &dto.RoleListResp{List: out}, nil
}

func (s *RoleService) Create(ctx context.Context, req dto.RoleCreateReq) error {
	role := &model.SysRole{
		ID:      uuid.NewString(),
		Code:    req.Code,
		Name:    req.Name,
		Status:  1,
		Deleted: 0,
	}
	return s.roleRepo.Create(ctx, role)
}

func (s *RoleService) Update(ctx context.Context, req dto.RoleUpdateReq) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["status"] = req.Status
	return s.roleRepo.UpdateByID(ctx, req.ID, updates)
}

func (s *RoleService) SetRoutes(ctx context.Context, req dto.RoleSetRoutesReq) error {
	if _, err := s.roleRepo.FindByID(ctx, req.RoleID); err != nil {
		return err
	}
	return s.roleRouteRepo.ReplaceRoleRoutes(ctx, req.RoleID, req.RouteIDs)
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Code == "admin" {
		return ErrReservedRole
	}
	n, err := s.userRoleRepo.CountUsersByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrRoleInUse
	}
	if err := s.roleRouteRepo.SoftDeleteByRoleID(ctx, id); err != nil {
		return err
	}
	return s.roleRepo.SoftDelete(ctx, id)
}
