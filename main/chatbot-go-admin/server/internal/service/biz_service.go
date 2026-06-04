package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chatbot-go-admin/server/internal/dto"
	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BizService struct {
	pkgRepo   *repository.BizPackageRepository
	quotaRepo *repository.BizUserQuotaRepository
	usageRepo *repository.BizUsageLogRepository
	annRepo   *repository.BizAnnouncementRepository
}

func NewBizService(
	pkgRepo *repository.BizPackageRepository,
	quotaRepo *repository.BizUserQuotaRepository,
	usageRepo *repository.BizUsageLogRepository,
	annRepo *repository.BizAnnouncementRepository,
) *BizService {
	return &BizService{
		pkgRepo:   pkgRepo,
		quotaRepo: quotaRepo,
		usageRepo: usageRepo,
		annRepo:   annRepo,
	}
}

func (s *BizService) QuotaSummary(ctx context.Context, userID string, isAdmin bool) (*dto.BizQuotaSummary, error) {
	if isAdmin {
		return &dto.BizQuotaSummary{
			IsUnlimited:    true,
			PackageTitle:   "管理员",
			TotalQuota:     0,
			UsedQuota:      0,
			RemainingQuota: -1,
		}, nil
	}
	q, err := s.quotaRepo.GetByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.BizQuotaSummary{
			PackageTitle:   "",
			TotalQuota:     0,
			UsedQuota:      0,
			RemainingQuota: 0,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	rem := q.TotalQuota - q.UsedQuota
	if rem < 0 {
		rem = 0
	}
	return &dto.BizQuotaSummary{
		PackageID:      q.PackageID,
		PackageTitle:   q.PackageTitle,
		TotalQuota:     q.TotalQuota,
		UsedQuota:      q.UsedQuota,
		RemainingQuota: rem,
	}, nil
}

func (s *BizService) UsageTrend(ctx context.Context, userID string, isAdmin bool, days int) ([]dto.BizUsageTrendPoint, error) {
	if days <= 0 {
		days = 14
	}
	if days > 90 {
		days = 90
	}
	since := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	counts, err := s.usageRepo.DailyCountsSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	byDay := map[string]int64{}
	for _, c := range counts {
		byDay[c.Day] = c.Count
	}
	out := make([]dto.BizUsageTrendPoint, 0, days)
	for i := 0; i < days; i++ {
		d := since.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		n := byDay[key]
		if isAdmin && n == 0 {
			// light demo curve for admin empty data (optional flat zero)
		}
		out = append(out, dto.BizUsageTrendPoint{Date: key, Count: n})
	}
	_ = isAdmin
	return out, nil
}

func (s *BizService) UsageList(ctx context.Context, userID string, page, pageSize int) (*dto.BizUsageListResp, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	rows, total, err := s.usageRepo.ListByUser(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.BizUsageLogItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, dto.BizUsageLogItem{
			ID:         r.ID,
			Channel:    r.Channel,
			RefText:    r.RefText,
			CreateTime: r.CreateTime,
		})
	}
	return &dto.BizUsageListResp{Total: total, List: list}, nil
}

func (s *BizService) Announcements(ctx context.Context, limit int) ([]dto.BizAnnouncementItem, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.annRepo.ListPublished(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BizAnnouncementItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.BizAnnouncementItem{
			ID:          r.ID,
			Title:       r.Title,
			Content:     r.Content,
			PublishTime: r.PublishTime,
		})
	}
	return out, nil
}

func (s *BizService) PackageCatalog(ctx context.Context) ([]dto.BizPackageCatalogItem, error) {
	rows, err := s.pkgRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BizPackageCatalogItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.BizPackageCatalogItem{
			ID:          r.ID,
			Code:        r.Code,
			Name:        r.Name,
			Description: r.Description,
			QuotaCalls:  r.QuotaCalls,
		})
	}
	return out, nil
}

// SimulatedPurchase applies package quota immediately (mock payment).
func (s *BizService) SimulatedPurchase(ctx context.Context, userID string, req dto.BizPurchaseReq) (*dto.BizPurchaseResp, error) {
	pkg, err := s.pkgRepo.FindByID(ctx, req.PackageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPackageNotFound
		}
		return nil, err
	}
	if pkg.Status != 1 {
		return nil, ErrPackageDisabled
	}
	now := time.Now()
	row := &model.BizUserQuota{
		UserID:       userID,
		PackageID:    pkg.ID,
		PackageTitle: pkg.Name,
		TotalQuota:   pkg.QuotaCalls,
		UsedQuota:    0,
		UpdateTime:   now,
	}
	if err := s.quotaRepo.Upsert(ctx, row); err != nil {
		return nil, err
	}
	rem := pkg.QuotaCalls
	return &dto.BizPurchaseResp{
		OK:             true,
		PackageTitle:   pkg.Name,
		TotalQuota:     pkg.QuotaCalls,
		UsedQuota:      0,
		RemainingQuota: rem,
		Message:        "模拟支付成功，套餐已生效",
	}, nil
}

// SimulateExperienceCall consumes one call and writes a log (for demo / experience chain).
func (s *BizService) SimulateExperienceCall(ctx context.Context, userID string, isAdmin bool) error {
	ref := fmt.Sprintf("experience-%s", time.Now().Format("15:04:05"))
	logRow := &model.BizUsageLog{
		ID:         uuid.NewString(),
		UserID:     userID,
		Channel:    "experience",
		RefText:    ref,
		CreateTime: time.Now(),
	}
	if isAdmin {
		return s.usageRepo.Create(ctx, logRow)
	}
	q, err := s.quotaRepo.GetByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 兼容旧用户：若没有额度记录，自动发放默认体验额度，避免首次体验接口直接 400。
		const defaultTrialQuota int64 = 20
		if err := s.quotaRepo.Upsert(ctx, &model.BizUserQuota{
			UserID:       userID,
			PackageID:    "",
			PackageTitle: "体验额度",
			TotalQuota:   defaultTrialQuota,
			UsedQuota:    0,
			UpdateTime:   time.Now(),
		}); err != nil {
			return err
		}
		q, err = s.quotaRepo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	if q.TotalQuota <= 0 {
		return ErrInsufficientQuota
	}
	if q.UsedQuota >= q.TotalQuota {
		return ErrInsufficientQuota
	}
	if _, err := s.quotaRepo.IncrementUsed(ctx, userID, 1); err != nil {
		return err
	}
	return s.usageRepo.Create(ctx, logRow)
}

// --- Admin packages ---

func (s *BizService) AdminPackageList(ctx context.Context) ([]dto.SysPackageItem, error) {
	rows, err := s.pkgRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SysPackageItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.SysPackageItem{
			ID:          r.ID,
			Code:        r.Code,
			Name:        r.Name,
			Description: r.Description,
			QuotaCalls:  r.QuotaCalls,
			SortNo:      r.SortNo,
			Status:      r.Status,
		})
	}
	return out, nil
}

func (s *BizService) AdminPackageCreate(ctx context.Context, req dto.SysPackageCreateReq) error {
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	now := time.Now()
	row := &model.BizPackage{
		ID:          uuid.NewString(),
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		QuotaCalls:  req.QuotaCalls,
		SortNo:      req.SortNo,
		Status:      status,
		CreateTime:  now,
		UpdateTime:  now,
		Deleted:     0,
	}
	return s.pkgRepo.Create(ctx, row)
}

func (s *BizService) AdminPackageUpdate(ctx context.Context, req dto.SysPackageUpdateReq) error {
	updates := map[string]any{}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.QuotaCalls != nil {
		updates["quota_calls"] = *req.QuotaCalls
	}
	if req.SortNo != nil {
		updates["sort_no"] = *req.SortNo
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		return nil
	}
	return s.pkgRepo.UpdateByID(ctx, req.ID, updates)
}

func (s *BizService) AdminPackageDelete(ctx context.Context, id string) error {
	return s.pkgRepo.SoftDelete(ctx, id)
}

// --- Admin announcements ---

func (s *BizService) AdminAnnouncementList(ctx context.Context) ([]dto.SysAnnouncementItem, error) {
	rows, err := s.annRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SysAnnouncementItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.SysAnnouncementItem{
			ID:          r.ID,
			Title:       r.Title,
			Content:     r.Content,
			SortNo:      r.SortNo,
			Status:      r.Status,
			PublishTime: r.PublishTime,
		})
	}
	return out, nil
}

func (s *BizService) AdminAnnouncementCreate(ctx context.Context, req dto.SysAnnouncementCreateReq) error {
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	now := time.Now()
	pt := req.PublishTime
	if pt == nil {
		pt = &now
	}
	row := &model.BizAnnouncement{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Content:     req.Content,
		SortNo:      req.SortNo,
		Status:      status,
		PublishTime: pt,
		CreateTime:  now,
		UpdateTime:  now,
		Deleted:     0,
	}
	return s.annRepo.Create(ctx, row)
}

func (s *BizService) AdminAnnouncementUpdate(ctx context.Context, req dto.SysAnnouncementUpdateReq) error {
	updates := map[string]any{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.SortNo != nil {
		updates["sort_no"] = *req.SortNo
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.PublishTime != nil {
		updates["publish_time"] = req.PublishTime
	}
	if len(updates) == 0 {
		return nil
	}
	return s.annRepo.UpdateByID(ctx, req.ID, updates)
}

func (s *BizService) AdminAnnouncementDelete(ctx context.Context, id string) error {
	return s.annRepo.SoftDelete(ctx, id)
}
