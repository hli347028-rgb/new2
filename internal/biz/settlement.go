package biz

import (
	"context"
	"fmt"
	"time"

	"backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

const (
	SettlementStatusRunning   = "running"
	SettlementStatusCompleted = "success"
	SettlementStatusSuccess   = "success"
	SettlementStatusFailed    = "failed"
)

// SettlementUsecase AIX daily settlement
type SettlementUsecase struct {
	userRepo    UserRepo
	stakingRepo StakingRepo
	walletRepo  WalletRepo
	log         *log.Helper
}

func NewSettlementUsecase(userRepo UserRepo, stakingRepo StakingRepo, walletRepo WalletRepo, logger log.Logger) *SettlementUsecase {
	return &SettlementUsecase{
		userRepo:    userRepo,
		stakingRepo: stakingRepo,
		walletRepo:  walletRepo,
		log:         log.NewHelper(logger),
	}
}

func chinaDate(t time.Time) string {
	return t.In(token.ChinaLocation()).Format("2006-01-02")
}

// RunDailySettlement 自动任务：已完成则跳过
func (uc *SettlementUsecase) RunDailySettlement(ctx context.Context, settlementDate string) error {
	return uc.runDailySettlement(ctx, settlementDate, false)
}

// ForceDailySettlement 管理端强制再跑（若当日已成功则先跳过幂等；force 时仍允许失败重跑）
func (uc *SettlementUsecase) ForceDailySettlement(ctx context.Context, settlementDate string) error {
	return uc.runDailySettlement(ctx, settlementDate, true)
}

func (uc *SettlementUsecase) runDailySettlement(ctx context.Context, settlementDate string, force bool) error {
	if !force {
		completed, err := uc.stakingRepo.HasCompletedSettlement(ctx, settlementDate)
		if err != nil {
			return err
		}
		if completed {
			uc.log.Infof("settlement %s already completed, skip", settlementDate)
			return nil
		}
	} else {
		completed, err := uc.stakingRepo.HasCompletedSettlement(ctx, settlementDate)
		if err != nil {
			return err
		}
		if completed {
			uc.log.Infof("settlement %s already success (unique date), skip force", settlementDate)
			return nil
		}
	}

	price, err := uc.stakingRepo.GetAixPrice(ctx, settlementDate)
	if err != nil {
		return err
	}
	if price == "" || price == "0" {
		price = decimal.NewFromFloat(AixPriceInitial).String()
		_ = uc.stakingRepo.UpsertAixPrice(ctx, settlementDate, price, "settlement default")
	}
	priceDec, _ := decimal.NewFromString(price)

	started := time.Now()
	batch := &SettlementBatch{
		SettlementDate: settlementDate,
		AixPrice:       priceDec.String(),
		Status:         SettlementStatusRunning,
		StartedAt:      started,
	}
	if err := uc.stakingRepo.CreateSettlementBatch(ctx, batch); err != nil {
		return err
	}
	uc.log.Infof("settlement %s batch#%d started", settlementDate, batch.ID)

	staticCount, staticAmt, err := uc.processStatic(ctx, settlementDate, batch.ID, priceDec)
	if err != nil {
		_ = uc.stakingRepo.FinishSettlementBatch(ctx, batch.ID, SettlementStatusFailed, 0, "0", 0, "0", err.Error())
		return err
	}
	if err := uc.refreshMgmtLevels(ctx); err != nil {
		_ = uc.stakingRepo.FinishSettlementBatch(ctx, batch.ID, SettlementStatusFailed, staticCount, staticAmt.String(), 0, "0", err.Error())
		return err
	}
	mgmtCount, mgmtAmt, err := uc.processMgmtRewards(ctx, settlementDate, batch.ID, staticAmt)
	if err != nil {
		_ = uc.stakingRepo.FinishSettlementBatch(ctx, batch.ID, SettlementStatusFailed, staticCount, staticAmt.String(), 0, "0", err.Error())
		return err
	}
	// Management rewards may also complete orders. Refresh again so those
	// orders are removed from every upstream performance branch immediately.
	if err := uc.refreshMgmtLevels(ctx); err != nil {
		_ = uc.stakingRepo.FinishSettlementBatch(ctx, batch.ID, SettlementStatusFailed, staticCount, staticAmt.String(), mgmtCount, mgmtAmt.String(), err.Error())
		return err
	}
	return uc.stakingRepo.FinishSettlementBatch(ctx, batch.ID, SettlementStatusSuccess, staticCount, staticAmt.String(), mgmtCount, mgmtAmt.String(), "")
}

func (uc *SettlementUsecase) processStatic(ctx context.Context, date string, batchID int64, aixPrice decimal.Decimal) (int32, decimal.Decimal, error) {
	orders, err := uc.stakingRepo.ListActiveOrders(ctx)
	if err != nil {
		return 0, decimal.Zero, err
	}
	if !aixPrice.IsPositive() {
		aixPrice = decimal.NewFromInt(1)
	}
	rate := decimal.NewFromFloat(StaticRate).Div(decimal.NewFromInt(100)) // 0.5% → 0.005
	var count int32
	totalAix := decimal.Zero

	for _, order := range orders {
		exists, err := uc.stakingRepo.HasStaticReward(ctx, order.ID, date)
		if err != nil {
			return 0, decimal.Zero, err
		}
		if exists {
			continue
		}
		principal, _ := decimal.NewFromString(order.Principal)
		exitCap, _ := decimal.NewFromString(order.ExitCap)
		earned, _ := decimal.NewFromString(order.EarnedTotal)
		remain := exitCap.Sub(earned)
		if remain.LessThanOrEqual(decimal.Zero) {
			now := time.Now()
			_ = uc.stakingRepo.UpdateOrderEarned(ctx, order.ID, earned.String(), OrderStatusExited, &now)
			continue
		}
		usdtValue := principal.Mul(rate)
		payUsdt := usdtValue
		if payUsdt.GreaterThan(remain) {
			payUsdt = remain
		}
		// aix_price = 每枚 AIX 的 USDT 价：价=1 → 1:1；价=2 → USDT:AIX=2:1
		aixAmount := payUsdt.Div(aixPrice)
		newEarned := earned.Add(payUsdt)
		status := OrderStatusActive
		var exitedAt *time.Time
		if newEarned.GreaterThanOrEqual(exitCap) {
			status = OrderStatusExited
			t := time.Now()
			exitedAt = &t
			newEarned = exitCap
		}
		// 静态：金本位 USDT 计入出局与静态总收益；折算 AIX 入代币余额
		if err := uc.creditStatic(ctx, order.UserID, aixAmount, payUsdt); err != nil {
			return 0, decimal.Zero, err
		}
		if err := uc.stakingRepo.UpdateOrderEarned(ctx, order.ID, newEarned.String(), status, exitedAt); err != nil {
			return 0, decimal.Zero, err
		}
		oid, bid := order.ID, batchID
		base := usdtValue
		rateDec := rate
		if err := uc.stakingRepo.CreateRewardLog(ctx, &RewardLog{
			UserID:         order.UserID,
			OrderID:        &oid,
			BatchID:        &bid,
			Type:           RewardTypeStaticAix,
			Asset:          TokenAIX,
			Amount:         aixAmount.String(),
			BaseAmount:     base.String(),
			Rate:           rateDec.String(),
			ExitApplied:    payUsdt.String(),
			SettlementDate: date,
		}); err != nil {
			return 0, decimal.Zero, err
		}
		count++
		totalAix = totalAix.Add(aixAmount)
	}
	return count, totalAix, nil
}

func (uc *SettlementUsecase) creditStatic(ctx context.Context, userID int64, aixAmount, usdtGold decimal.Decimal) error {
	if aixAmount.LessThanOrEqual(decimal.Zero) && usdtGold.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return err
	}
	aixCur, _ := decimal.NewFromString(user.AixBalance)
	usdtCur, _ := decimal.NewFromString(user.StaticUsdtTotal)
	upd := &AdminUserUpdate{UserID: userID}
	if aixAmount.GreaterThan(decimal.Zero) {
		upd.AixBalance = aixCur.Add(aixAmount).String()
	}
	if usdtGold.GreaterThan(decimal.Zero) {
		upd.StaticUsdtTotal = usdtCur.Add(usdtGold).String()
	}
	return uc.userRepo.AdminUpdateUser(ctx, upd)
}

func (uc *SettlementUsecase) creditUsdtReward(ctx context.Context, userID int64, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return err
	}
	cur, _ := decimal.NewFromString(user.UsdtReward)
	return uc.userRepo.AdminUpdateUser(ctx, &AdminUserUpdate{
		UserID:     userID,
		UsdtReward: cur.Add(amount).String(),
	})
}

// refreshMgmtLevels 刷新全网小区业绩与 W 等级
func (uc *SettlementUsecase) refreshMgmtLevels(ctx context.Context) error {
	return uc.userRepo.RefreshPerformance(ctx)
}

// processMgmtRewards 级差管理奖：以当日静态 USDT 池为底数，按上级链级差分配
func (uc *SettlementUsecase) processMgmtRewards(ctx context.Context, date string, batchID int64, staticAixTotal decimal.Decimal) (int32, decimal.Decimal, error) {
	// 底数：当日静态金本位合计 ≈ static_aix * price；用 reward_logs 汇总 exit_applied 更准
	basePool, err := uc.stakingRepo.SumStaticByDate(ctx, date)
	if err != nil {
		return 0, decimal.Zero, err
	}
	pool, _ := decimal.NewFromString(basePool)
	if pool.LessThanOrEqual(decimal.Zero) {
		// fallback: aix total * 1
		pool = staticAixTotal
	}
	if pool.LessThanOrEqual(decimal.Zero) {
		return 0, decimal.Zero, nil
	}

	users, err := uc.userRepo.ListAllUsers(ctx)
	if err != nil {
		return 0, decimal.Zero, err
	}
	byID := map[int64]*User{}
	for _, u := range users {
		byID[u.ID] = u
	}

	var count int32
	total := decimal.Zero
	// 对每个有静态产出的用户，沿上级链发级差
	sourceIDs, err := uc.stakingRepo.ListUserIDsWithReleaseOnDate(ctx, date)
	if err != nil {
		return 0, decimal.Zero, err
	}
	for _, sid := range sourceIDs {
		srcAmtStr, err := uc.stakingRepo.SumReleaseByUserDate(ctx, sid, date)
		if err != nil {
			return 0, decimal.Zero, err
		}
		srcBase, _ := decimal.NewFromString(srcAmtStr)
		if srcBase.LessThanOrEqual(decimal.Zero) {
			continue
		}
		src := byID[sid]
		if src == nil {
			continue
		}
		lastRate := 0.0
		current := src
		seen := map[int64]bool{sid: true}
		for current.InviterID != nil {
			pid := *current.InviterID
			if seen[pid] {
				break
			}
			seen[pid] = true
			parent := byID[pid]
			if parent == nil {
				break
			}
			rate := MgmtRateForLevel(parent.MgmtLevel)
			diff := rate - lastRate
			if diff > 0 {
				pay := srcBase.Mul(decimal.NewFromFloat(diff))
				pay, applied, err := uc.applyExitCapReward(ctx, parent.ID, pay)
				if err != nil {
					return 0, decimal.Zero, err
				}
				if pay.GreaterThan(decimal.Zero) {
					if err := uc.creditUsdtReward(ctx, parent.ID, pay); err != nil {
						return 0, decimal.Zero, err
					}
					oid := parent.ID
					_ = oid
					bid := batchID
					fromID := sid
					rateDec := decimal.NewFromFloat(diff)
					base := srcBase
					if err := uc.stakingRepo.CreateRewardLog(ctx, &RewardLog{
						UserID:         parent.ID,
						FromUserID:     &fromID,
						BatchID:        &bid,
						Type:           RewardTypeMgmt,
						Asset:          TokenUSDT,
						Amount:         pay.String(),
						BaseAmount:     base.String(),
						Rate:           rateDec.String(),
						ExitApplied:    applied.String(),
						SettlementDate: date,
					}); err != nil {
						return 0, decimal.Zero, err
					}
					count++
					total = total.Add(pay)
				}
				lastRate = rate
			} else if rate > lastRate {
				lastRate = rate
			}
			current = parent
		}
	}
	return count, total, nil
}

// applyExitCapReward 将奖励计入活跃订单出局进度，返回实发与计入出局金额
func (uc *SettlementUsecase) applyExitCapReward(ctx context.Context, userID int64, want decimal.Decimal) (pay, exitApplied decimal.Decimal, err error) {
	if want.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, nil
	}
	orders, err := uc.stakingRepo.ListActiveOrdersByUser(ctx, userID)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	remainTotal := decimal.Zero
	for _, o := range orders {
		cap, _ := decimal.NewFromString(o.ExitCap)
		earned, _ := decimal.NewFromString(o.EarnedTotal)
		r := cap.Sub(earned)
		if r.IsPositive() {
			remainTotal = remainTotal.Add(r)
		}
	}
	pay = want
	if remainTotal.LessThanOrEqual(decimal.Zero) {
		// 无活跃订单：仍可发奖励但不计入出局
		if MgmtCountsTowardExit {
			return decimal.Zero, decimal.Zero, nil
		}
		return pay, decimal.Zero, nil
	}
	if pay.GreaterThan(remainTotal) {
		pay = remainTotal
	}
	exitApplied = decimal.Zero
	if MgmtCountsTowardExit {
		left := pay
		for _, o := range orders {
			if left.LessThanOrEqual(decimal.Zero) {
				break
			}
			cap, _ := decimal.NewFromString(o.ExitCap)
			earned, _ := decimal.NewFromString(o.EarnedTotal)
			remain := cap.Sub(earned)
			if remain.LessThanOrEqual(decimal.Zero) {
				continue
			}
			apply := left
			if apply.GreaterThan(remain) {
				apply = remain
			}
			newEarned := earned.Add(apply)
			status := OrderStatusActive
			var exitedAt *time.Time
			if newEarned.GreaterThanOrEqual(cap) {
				status = OrderStatusExited
				t := time.Now()
				exitedAt = &t
				newEarned = cap
			}
			if err := uc.stakingRepo.UpdateOrderEarned(ctx, o.ID, newEarned.String(), status, exitedAt); err != nil {
				return decimal.Zero, decimal.Zero, err
			}
			exitApplied = exitApplied.Add(apply)
			left = left.Sub(apply)
		}
	}
	return pay, exitApplied, nil
}

func (uc *SettlementUsecase) ListBatches(ctx context.Context, offset, limit int) ([]*SettlementBatch, int64, error) {
	return uc.stakingRepo.ListSettlementBatches(ctx, offset, limit)
}

func (uc *SettlementUsecase) SumReleaseByDate(ctx context.Context, date string) (string, error) {
	return uc.stakingRepo.SumReleaseByDate(ctx, date)
}

func (uc *SettlementUsecase) SumReleaseForBatch(ctx context.Context, b *SettlementBatch) (string, error) {
	if b == nil {
		return "0", nil
	}
	if b.StaticAmount != "" {
		return b.StaticAmount, nil
	}
	return uc.stakingRepo.SumReleaseByDate(ctx, b.SettlementDate)
}

// TodaySettlementDate 返回当前中国时区应对应的结算日期（昨日）
func TodaySettlementDate(now time.Time) string {
	now = now.In(token.ChinaLocation())
	return chinaDate(now.Add(-24 * time.Hour))
}

// BackfillMissingEcoRewards disabled in AIX
func (uc *SettlementUsecase) BackfillMissingEcoRewards(ctx context.Context) error {
	return fmt.Errorf("not supported in AIX")
}
