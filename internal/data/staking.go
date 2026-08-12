package data

import (
	"context"
	"time"

	"backend/internal/biz"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type stakingRepo struct {
	data *Data
}

func NewStakingRepo(data *Data) biz.StakingRepo {
	return &stakingRepo{data: data}
}

func (r *stakingRepo) ListActiveOrders(ctx context.Context) ([]*biz.StakingOrder, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).Where("status = ?", biz.OrderStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	return r.toStaking(list), nil
}

func (r *stakingRepo) ListActiveOrdersByUser(ctx context.Context, userID int64) ([]*biz.StakingOrder, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, biz.OrderStatusActive).
		Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return r.toStaking(list), nil
}

func (r *stakingRepo) ListOrdersByUser(ctx context.Context, userID int64) ([]*biz.StakingOrder, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return r.toStaking(list), nil
}

func (r *stakingRepo) toStaking(list []OrderPO) []*biz.StakingOrder {
	out := make([]*biz.StakingOrder, 0, len(list))
	for _, po := range list {
		out = append(out, &biz.StakingOrder{
			ID:          po.ID,
			UserID:      po.UserID,
			Principal:   po.Principal.String(),
			ExitCap:     po.ExitCap.String(),
			EarnedTotal: po.EarnedTotal.String(),
			Status:      po.Status,
			FundSource:  po.FundSource,
			CreatedAt:   po.CreatedTime,
		})
	}
	return out
}

func (r *stakingRepo) UpdateOrderEarned(ctx context.Context, orderID int64, earnedTotal, status string, exitedTime *time.Time) error {
	earned, err := decimal.NewFromString(earnedTotal)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"earned_total": earned,
		"status":       status,
	}
	if exitedTime != nil {
		updates["exited_time"] = *exitedTime
	}
	return r.data.db.WithContext(ctx).Model(&OrderPO{}).Where("id = ?", orderID).Updates(updates).Error
}

func (r *stakingRepo) CreateRewardLog(ctx context.Context, log *biz.RewardLog) error {
	return NewWalletRepo(r.data).CreateRewardLog(ctx, log)
}

func (r *stakingRepo) ListRewardLogsByUser(ctx context.Context, userID int64) ([]*biz.RewardLog, error) {
	return NewWalletRepo(r.data).ListRewardLogsByUser(ctx, userID)
}

func (r *stakingRepo) ListStaticRewardsAsRelease(ctx context.Context, userID int64) ([]*biz.ReleaseRecord, error) {
	var list []RewardLogPO
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, biz.RewardTypeStaticAix).
		Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.ReleaseRecord, 0, len(list))
	for _, po := range list {
		oid := int64(0)
		if po.OrderID != nil {
			oid = *po.OrderID
		}
		date := ""
		if po.SettlementDate != nil {
			date = *po.SettlementDate
		}
		rate := ""
		if po.Rate != nil {
			rate = po.Rate.String()
		}
		out = append(out, &biz.ReleaseRecord{
			ID:             po.ID,
			UserID:         po.UserID,
			OrderID:        oid,
			SettlementDate: date,
			Rate:           rate,
			Amount:         po.Amount.String(),
			CreatedAt:      po.CreatedTime,
		})
	}
	return out, nil
}

func (r *stakingRepo) ListDynamicRewardsAsReferral(ctx context.Context, userID int64) ([]*biz.ReferralReward, error) {
	var list []RewardLogPO
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, biz.RewardTypeDynamicUsdt).
		Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.ReferralReward, 0, len(list))
	for _, po := range list {
		src := int64(0)
		if po.FromUserID != nil {
			src = *po.FromUserID
		}
		oid := int64(0)
		if po.OrderID != nil {
			oid = *po.OrderID
		}
		date := ""
		if po.SettlementDate != nil {
			date = *po.SettlementDate
		}
		base, rate := "", ""
		if po.BaseAmount != nil {
			base = po.BaseAmount.String()
		}
		if po.Rate != nil {
			rate = po.Rate.String()
		}
		out = append(out, &biz.ReferralReward{
			ID:                po.ID,
			BeneficiaryUserID: po.UserID,
			SourceUserID:      src,
			SourceOrderID:     oid,
			Generation:        1,
			BaseAmount:        base,
			Rate:              rate,
			RewardAmount:      po.Amount.String(),
			SettlementDate:    date,
			CreatedAt:         po.CreatedTime,
		})
	}
	return out, nil
}

func (r *stakingRepo) HasStaticReward(ctx context.Context, orderID int64, date string) (bool, error) {
	var cnt int64
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Where("order_id = ? AND type = ? AND settlement_date = ?", orderID, biz.RewardTypeStaticAix, date).
		Count(&cnt).Error
	return cnt > 0, err
}

func (r *stakingRepo) GetAixPrice(ctx context.Context, date string) (string, error) {
	return NewWalletRepo(r.data).GetAixPrice(ctx, date)
}

func (r *stakingRepo) UpsertAixPrice(ctx context.Context, date, price, remark string) error {
	return NewWalletRepo(r.data).UpsertAixPrice(ctx, date, price, remark)
}

func (r *stakingRepo) HasCompletedSettlement(ctx context.Context, date string) (bool, error) {
	var cnt int64
	err := r.data.db.WithContext(ctx).Model(&SettlementBatchPO{}).
		Where("settlement_date = ? AND status = ?", date, biz.SettlementStatusSuccess).
		Count(&cnt).Error
	return cnt > 0, err
}

func (r *stakingRepo) CreateSettlementBatch(ctx context.Context, batch *biz.SettlementBatch) error {
	price, _ := decimal.NewFromString(batch.AixPrice)
	started := batch.StartedAt
	po := &SettlementBatchPO{
		SettlementDate: batch.SettlementDate,
		AixPrice:       price,
		Status:         batch.Status,
		StartedTime:    &started,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	batch.ID = po.ID
	return nil
}

func (r *stakingRepo) FinishSettlementBatch(ctx context.Context, id int64, status string, staticCount int32, staticAmount string, mgmtCount int32, mgmtAmount string, errMsg string) error {
	sa, _ := decimal.NewFromString(staticAmount)
	ma, _ := decimal.NewFromString(mgmtAmount)
	now := time.Now()
	return r.data.db.WithContext(ctx).Model(&SettlementBatchPO{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"static_count":  staticCount,
		"static_amount": sa,
		"mgmt_count":    mgmtCount,
		"mgmt_amount":   ma,
		"finished_time": now,
		"error_msg":     errMsg,
	}).Error
}

func (r *stakingRepo) ListSettlementBatches(ctx context.Context, offset, limit int) ([]*biz.SettlementBatch, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int64
	if err := r.data.db.WithContext(ctx).Model(&SettlementBatchPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SettlementBatchPO
	if err := r.data.db.WithContext(ctx).Order("id desc").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*biz.SettlementBatch, 0, len(list))
	for _, po := range list {
		b := &biz.SettlementBatch{
			ID:             po.ID,
			SettlementDate: po.SettlementDate,
			AixPrice:       po.AixPrice.String(),
			Status:         po.Status,
			StaticCount:    po.StaticCount,
			StaticAmount:   po.StaticAmount.String(),
			MgmtCount:      po.MgmtCount,
			MgmtAmount:     po.MgmtAmount.String(),
			ErrorMsg:       po.ErrorMsg,
			CreatedTime:    po.CreatedTime,
		}
		if po.StartedTime != nil {
			b.StartedAt = *po.StartedTime
		}
		b.FinishedAt = po.FinishedTime
		out = append(out, b)
	}
	return out, total, nil
}

func (r *stakingRepo) SumStaticByDate(ctx context.Context, date string) (string, error) {
	type row struct{ Total decimal.Decimal }
	var res row
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Select("COALESCE(SUM(exit_applied),0) as total").
		Where("type = ? AND settlement_date = ?", biz.RewardTypeStaticAix, date).
		Scan(&res).Error
	if err != nil {
		return "0", err
	}
	return res.Total.String(), nil
}

func (r *stakingRepo) UpdateOrderAfterRelease(ctx context.Context, orderID int64, releasedAmount, status string, releaseDay, rateIndex int32, rateGoingUp bool) error {
	if status == "completed" {
		status = biz.OrderStatusExited
	}
	return r.UpdateOrderEarned(ctx, orderID, releasedAmount, status, nil)
}

func (r *stakingRepo) CreateReleaseRecord(ctx context.Context, record *biz.ReleaseRecord) error {
	return nil
}

func (r *stakingRepo) ListReleaseRecordsByUser(ctx context.Context, userID int64) ([]*biz.ReleaseRecord, error) {
	return r.ListStaticRewardsAsRelease(ctx, userID)
}

func (r *stakingRepo) CreateReferralReward(ctx context.Context, reward *biz.ReferralReward) error {
	return nil
}

func (r *stakingRepo) ListReferralRewardsByUser(ctx context.Context, userID int64) ([]*biz.ReferralReward, error) {
	return r.ListDynamicRewardsAsReferral(ctx, userID)
}

func (r *stakingRepo) SumReferralByOrderDate(ctx context.Context, orderID int64, settlementDate string) (string, error) {
	type row struct{ Total decimal.Decimal }
	var res row
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Select("COALESCE(SUM(amount),0) as total").
		Where("order_id = ? AND type = ?", orderID, biz.RewardTypeDynamicUsdt).
		Scan(&res).Error
	return res.Total.String(), err
}

func (r *stakingRepo) CreateEcoReward(ctx context.Context, reward *biz.EcoReward) error {
	return nil
}
func (r *stakingRepo) ListEcoRewardsByUser(ctx context.Context, userID int64) ([]*biz.EcoReward, error) {
	return []*biz.EcoReward{}, nil
}
func (r *stakingRepo) CountEcoRewardsByDate(ctx context.Context, date string) (int64, error) {
	return 0, nil
}
func (r *stakingRepo) HasEcoReward(ctx context.Context, userID int64, date string) (bool, error) {
	return false, nil
}
func (r *stakingRepo) ListReleaseSettlementDates(ctx context.Context) ([]string, error) {
	var dates []string
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Where("type = ? AND settlement_date IS NOT NULL", biz.RewardTypeStaticAix).
		Distinct().Pluck("settlement_date", &dates).Error
	return dates, err
}
func (r *stakingRepo) GetLatestSettlementBatch(ctx context.Context, date string) (*biz.SettlementBatch, error) {
	var po SettlementBatchPO
	err := r.data.db.WithContext(ctx).Where("settlement_date = ?", date).Order("id desc").First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b := &biz.SettlementBatch{
		ID: po.ID, SettlementDate: po.SettlementDate, AixPrice: po.AixPrice.String(),
		Status: po.Status, StaticCount: po.StaticCount, StaticAmount: po.StaticAmount.String(),
		MgmtCount: po.MgmtCount, MgmtAmount: po.MgmtAmount.String(), ErrorMsg: po.ErrorMsg,
		CreatedTime: po.CreatedTime,
	}
	if po.StartedTime != nil {
		b.StartedAt = *po.StartedTime
	}
	b.FinishedAt = po.FinishedTime
	return b, nil
}

func (r *stakingRepo) SumReleaseByUserDate(ctx context.Context, userID int64, date string) (string, error) {
	type row struct{ Total decimal.Decimal }
	var res row
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Select("COALESCE(SUM(exit_applied),0) as total").
		Where("user_id = ? AND type = ? AND settlement_date = ?", userID, biz.RewardTypeStaticAix, date).
		Scan(&res).Error
	return res.Total.String(), err
}

func (r *stakingRepo) SumReleaseByUserDateSince(ctx context.Context, userID int64, date string, since time.Time) (string, error) {
	return r.SumReleaseByUserDate(ctx, userID, date)
}

func (r *stakingRepo) SumReleaseByDate(ctx context.Context, date string) (string, error) {
	return r.SumStaticByDate(ctx, date)
}

func (r *stakingRepo) SumReleaseForBatch(ctx context.Context, date string, startedAt time.Time, finishedAt *time.Time) (string, error) {
	return r.SumStaticByDate(ctx, date)
}

func (r *stakingRepo) ListUserIDsWithReleaseOnDate(ctx context.Context, date string) ([]int64, error) {
	var ids []int64
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Where("type = ? AND settlement_date = ?", biz.RewardTypeStaticAix, date).
		Distinct().Pluck("user_id", &ids).Error
	return ids, err
}

func (r *stakingRepo) ListUserIDsWithReleaseOnDateSince(ctx context.Context, date string, since time.Time) ([]int64, error) {
	return r.ListUserIDsWithReleaseOnDate(ctx, date)
}

func (r *stakingRepo) ListUserIDsWithStaticByBatch(ctx context.Context, batchID int64) ([]int64, error) {
	var ids []int64
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Where("type = ? AND batch_id = ?", biz.RewardTypeStaticAix, batchID).
		Distinct().Pluck("user_id", &ids).Error
	return ids, err
}

func (r *stakingRepo) SumStaticByUserBatch(ctx context.Context, userID, batchID int64) (string, error) {
	type row struct{ Total decimal.Decimal }
	var result row
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Select("COALESCE(SUM(exit_applied),0) AS total").
		Where("user_id = ? AND type = ? AND batch_id = ?", userID, biz.RewardTypeStaticAix, batchID).
		Scan(&result).Error
	return result.Total.String(), err
}

func (r *stakingRepo) SumSettledByUser(ctx context.Context, userID int64) (string, error) {
	type row struct{ Total decimal.Decimal }
	var res row
	err := r.data.db.WithContext(ctx).Model(&RewardLogPO{}).
		Select("COALESCE(SUM(amount),0) as total").
		Where("user_id = ?", userID).Scan(&res).Error
	return res.Total.String(), err
}
