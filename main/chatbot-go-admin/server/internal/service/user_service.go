package service

import (
	"context"
	"errors"
	"strings"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"chatbot-go-admin/server/internal/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultMemberRoleID 预置「普通用户」角色（含首页、套餐购买等路由）；与 resources/schema/sys_role.sql 一致。
const DefaultMemberRoleID = "00000000-0000-0000-0000-000000000003"

type UserService struct {
	userRepo     *repository.UserRepository
	userRoleRepo *repository.UserRoleRepository
	quotaRepo    *repository.BizUserQuotaRepository
	pkgRepo      *repository.BizPackageRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
	userRoleRepo *repository.UserRoleRepository,
	quotaRepo *repository.BizUserQuotaRepository,
	pkgRepo *repository.BizPackageRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		quotaRepo:    quotaRepo,
		pkgRepo:      pkgRepo,
	}
}

func (s *UserService) List(ctx context.Context, req dto.UserListReq) (*dto.UserListResp, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.List(ctx, strings.TrimSpace(req.Keyword), offset, pageSize)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	qm, err := s.quotaRepo.MapByUserIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]dto.UserItem, 0, len(users))
	for _, u := range users {
		rids, err := s.userRoleRepo.ListRoleIDsByUserID(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		item := dto.UserItem{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Status:   u.Status,
			IsAdmin:  u.IsAdmin,
			RoleIDs:  rids,
		}
		if q, ok := qm[u.ID]; ok {
			item.TotalQuota = q.TotalQuota
			item.UsedQuota = q.UsedQuota
			item.PackageTitle = q.PackageTitle
		}
		out = append(out, item)
	}
	return &dto.UserListResp{Total: total, List: out}, nil
}

func (s *UserService) Create(ctx context.Context, req dto.UserCreateReq) error {
	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if req.DefaultPackageID != "" {
		if _, err := s.pkgRepo.FindByID(ctx, req.DefaultPackageID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPackageNotFound
			}
			return err
		}
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return err
	}

	u := &model.SysUser{
		ID:           uuid.NewString(),
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Status:       1,
		IsAdmin:      req.IsAdmin,
		Deleted:      0,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return err
	}
	roleIDs := req.RoleIDs
	if len(roleIDs) == 0 && req.IsAdmin != 1 {
		roleIDs = []string{DefaultMemberRoleID}
	}
	if err := s.userRoleRepo.ReplaceUserRoles(ctx, u.ID, roleIDs); err != nil {
		return err
	}
	return s.initUserQuota(ctx, u.ID, req.TrialQuota, req.DefaultPackageID)
}

func (s *UserService) Update(ctx context.Context, req dto.UserUpdateReq) error {
	if req.DefaultPackageID != "" {
		if _, err := s.pkgRepo.FindByID(ctx, req.DefaultPackageID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPackageNotFound
			}
			return err
		}
	}

	updates := map[string]any{
		"nickname": req.Nickname,
		"status":   req.Status,
		"is_admin": req.IsAdmin,
	}
	if req.Password != "" {
		hash, err := util.HashPassword(req.Password)
		if err != nil {
			return err
		}
		updates["password_hash"] = hash
	}
	if err := s.userRepo.UpdateByID(ctx, req.ID, updates); err != nil {
		return err
	}
	if err := s.userRoleRepo.ReplaceUserRoles(ctx, req.ID, req.RoleIDs); err != nil {
		return err
	}
	if req.TrialQuota != nil || req.DefaultPackageID != "" {
		return s.applyQuotaAdmin(ctx, req.ID, req.TrialQuota, req.DefaultPackageID)
	}
	return nil
}

func (s *UserService) initUserQuota(ctx context.Context, userID string, trial *int64, packageID string) error {
	var total int64
	title := ""
	pkgID := ""
	if packageID != "" {
		p, err := s.pkgRepo.FindByID(ctx, packageID)
		if err != nil {
			return err
		}
		total = p.QuotaCalls
		title = p.Name
		pkgID = p.ID
	} else if trial != nil {
		total = *trial
		title = "体验额度"
	}
	row := &model.BizUserQuota{
		UserID:       userID,
		PackageID:    pkgID,
		PackageTitle: title,
		TotalQuota:   total,
		UsedQuota:    0,
	}
	return s.quotaRepo.Upsert(ctx, row)
}

func (s *UserService) applyQuotaAdmin(ctx context.Context, userID string, trial *int64, packageID string) error {
	var total int64
	title := ""
	pkgID := ""
	if packageID != "" {
		p, err := s.pkgRepo.FindByID(ctx, packageID)
		if err != nil {
			return err
		}
		total = p.QuotaCalls
		title = p.Name
		pkgID = p.ID
	} else if trial != nil {
		total = *trial
		title = "体验额度"
	}
	used := int64(0)
	if existing, err := s.quotaRepo.GetByUserID(ctx, userID); err == nil {
		used = existing.UsedQuota
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if total < used {
		used = total
	}
	row := &model.BizUserQuota{
		UserID:       userID,
		PackageID:    pkgID,
		PackageTitle: title,
		TotalQuota:   total,
		UsedQuota:    used,
	}
	return s.quotaRepo.Upsert(ctx, row)
}

func (s *UserService) Delete(ctx context.Context, operatorUserID, targetID string) error {
	if operatorUserID == targetID {
		return ErrCannotDeleteSelf
	}
	u, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if u.IsAdmin == 1 {
		n, err := s.userRepo.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if n <= 1 {
			return ErrLastAdminUser
		}
	}
	if err := s.userRoleRepo.SoftDeleteByUserID(ctx, targetID); err != nil {
		return err
	}
	return s.userRepo.SoftDelete(ctx, targetID)
}

