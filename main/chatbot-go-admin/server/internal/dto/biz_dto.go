package dto

import "time"

// --- User-facing ---

type BizQuotaSummary struct {
	IsUnlimited   bool   `json:"isUnlimited"`
	PackageID     string `json:"packageId"`
	PackageTitle  string `json:"packageTitle"`
	TotalQuota    int64  `json:"totalQuota"`
	UsedQuota     int64  `json:"usedQuota"`
	RemainingQuota int64 `json:"remainingQuota"`
}

type BizUsageTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type BizUsageLogItem struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"`
	RefText    string    `json:"refText"`
	CreateTime time.Time `json:"createTime"`
}

type BizUsageListResp struct {
	Total int64             `json:"total"`
	List  []BizUsageLogItem `json:"list"`
}

type BizAnnouncementItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	PublishTime *time.Time `json:"publishTime"`
}

type BizPackageCatalogItem struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	QuotaCalls  int64  `json:"quotaCalls"`
}

type BizPurchaseReq struct {
	PackageID string `json:"packageId" binding:"required"`
}

type BizPurchaseResp struct {
	OK            bool   `json:"ok"`
	PackageTitle  string `json:"packageTitle"`
	TotalQuota    int64  `json:"totalQuota"`
	UsedQuota     int64  `json:"usedQuota"`
	RemainingQuota int64 `json:"remainingQuota"`
	Message       string `json:"message"`
}

// --- Admin: package ---

type SysPackageItem struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	QuotaCalls  int64  `json:"quotaCalls"`
	SortNo      int    `json:"sortNo"`
	Status      int    `json:"status"`
}

type SysPackageCreateReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	QuotaCalls  int64  `json:"quotaCalls" binding:"required"`
	SortNo      int    `json:"sortNo"`
	Status      *int   `json:"status"`
}

type SysPackageUpdateReq struct {
	ID          string `json:"id" binding:"required"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	QuotaCalls  *int64 `json:"quotaCalls"`
	SortNo      *int   `json:"sortNo"`
	Status      *int   `json:"status"`
}

// --- Admin: announcement ---

type SysAnnouncementItem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	SortNo      int        `json:"sortNo"`
	Status      int        `json:"status"`
	PublishTime *time.Time `json:"publishTime"`
}

type SysAnnouncementCreateReq struct {
	Title       string     `json:"title" binding:"required"`
	Content     string     `json:"content" binding:"required"`
	SortNo      int        `json:"sortNo"`
	Status      *int       `json:"status"`
	PublishTime *time.Time `json:"publishTime"`
}

type SysAnnouncementUpdateReq struct {
	ID          string     `json:"id" binding:"required"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	SortNo      *int       `json:"sortNo"`
	Status      *int       `json:"status"`
	PublishTime *time.Time `json:"publishTime"`
}
