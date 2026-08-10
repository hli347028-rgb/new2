package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/biz"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepo struct {
	data *Data
}

func NewWalletRepo(data *Data) biz.WalletRepo {
	return &walletRepo{data: data}
}

func (r *walletRepo) CreateRecharge(ctx context.Context, recharge *biz.Recharge) (*biz.Recharge, error) {
	amount, err := decimal.NewFromString(recharge.Amount)
	if err != nil {
		return nil, err
	}
	// unique placeholder until confirm
	placeholder := fmt.Sprintf("pending-%d-%d", recharge.UserID, time.Now().UnixNano())
	expire := recharge.ExpireAt
	po := &RechargePO{
		UserID:      recharge.UserID,
		Amount:      amount,
		TxHash:      placeholder,
		FromAddress: recharge.FromAddress,
		ToAddress:   recharge.ToAddress,
		Status:      recharge.Status,
		Message:     recharge.Message,
		ExpireAt:    &expire,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return r.rechargeToBiz(po), nil
}

func (r *walletRepo) FindRecharge(ctx context.Context, id int64) (*biz.Recharge, error) {
	var po RechargePO
	err := r.data.db.WithContext(ctx).First(&po, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.rechargeToBiz(&po), nil
}

func (r *walletRepo) FindRechargeByTxHash(ctx context.Context, txHash string) (*biz.Recharge, error) {
	txHash = strings.TrimSpace(txHash)
	if txHash == "" {
		return nil, nil
	}
	var po RechargePO
	err := r.data.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.rechargeToBiz(&po), nil
}

func (r *walletRepo) ConfirmRecharge(ctx context.Context, id int64, txHash string) error {
	_, err := r.ConfirmRechargeCredit(ctx, id, txHash)
	return err
}

func (r *walletRepo) ConfirmRechargeCredit(ctx context.Context, id int64, txHash string) (string, error) {
	var newBal string
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po RechargePO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, id).Error; err != nil {
			return err
		}
		if po.Status == biz.RechargeStatusConfirmed {
			return fmt.Errorf("already confirmed")
		}
		now := time.Now()
		if err := tx.Model(&po).Updates(map[string]interface{}{
			"tx_hash":        txHash,
			"status":         biz.RechargeStatusConfirmed,
			"confirmed_time": now,
		}).Error; err != nil {
			return err
		}
		var user UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, po.UserID).Error; err != nil {
			return err
		}
		user.UsdtRecharge = user.UsdtRecharge.Add(po.Amount)
		newBal = user.UsdtRecharge.String()
		return tx.Model(&user).Update("usdt_recharge", user.UsdtRecharge).Error
	})
	return newBal, err
}

func (r *walletRepo) ListRechargesByUser(ctx context.Context, userID int64) ([]*biz.Recharge, error) {
	var list []RechargePO
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.Recharge, 0, len(list))
	for i := range list {
		out = append(out, r.rechargeToBiz(&list[i]))
	}
	return out, nil
}

func (r *walletRepo) Subscribe(ctx context.Context, userID int64, amount, payFrom string, exitMul, directRate float64) (*biz.Order, string, error) {
	principal, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, "", err
	}
	if exitMul <= 0 {
		exitMul = 4
	}
	if directRate <= 0 {
		directRate = 0.5
	}
	exitCap := principal.Mul(decimal.NewFromFloat(exitMul))
	var created *biz.Order
	var balOut string

	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		fromRecharge := decimal.Zero
		fromReward := decimal.Zero
		directBase := decimal.Zero
		switch payFrom {
		case biz.PayFromRecharge:
			if user.UsdtRecharge.LessThan(principal) {
				return fmt.Errorf("insufficient usdt_recharge")
			}
			user.UsdtRecharge = user.UsdtRecharge.Sub(principal)
			fromRecharge = principal
			directBase = principal
			balOut = user.UsdtRecharge.String()
		case biz.PayFromReward:
			if user.UsdtReward.LessThan(principal) {
				return fmt.Errorf("insufficient usdt_reward")
			}
			user.UsdtReward = user.UsdtReward.Sub(principal)
			fromReward = principal
			balOut = user.UsdtReward.String()
		default:
			return fmt.Errorf("invalid pay_from")
		}
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"usdt_recharge": user.UsdtRecharge,
			"usdt_reward":   user.UsdtReward,
		}).Error; err != nil {
			return err
		}
		po := &OrderPO{
			UserID:       userID,
			Principal:    principal,
			ExitCap:      exitCap,
			EarnedTotal:  decimal.Zero,
			DirectBase:   directBase,
			FromRecharge: fromRecharge,
			FromReward:   fromReward,
			FundSource:   payFrom,
			Status:       biz.OrderStatusActive,
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		created = r.orderToBiz(po)

		// 直推奖：仅 recharge 且有上级
		if directBase.IsPositive() && user.InviterID != nil {
			if err := r.payDirectReward(tx, *user.InviterID, userID, po.ID, directBase, directRate); err != nil {
				return err
			}
		}
		// Keep the order and every ancestor's cached performance atomic.
		if err := refreshAncestorPerformance(tx, userID); err != nil {
			return err
		}
		return nil
	})
	return created, balOut, err
}

func (r *walletRepo) payDirectReward(tx *gorm.DB, inviterID, fromUserID, orderID int64, directBase decimal.Decimal, directRate float64) error {
	want := directBase.Mul(decimal.NewFromFloat(directRate))
	var orders []OrderPO
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND status = ?", inviterID, biz.OrderStatusActive).
		Order("id asc").Find(&orders).Error; err != nil {
		return err
	}
	remainCap := decimal.Zero
	for _, o := range orders {
		remainCap = remainCap.Add(o.ExitCap.Sub(o.EarnedTotal))
	}
	if remainCap.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	pay := want
	if pay.GreaterThan(remainCap) {
		pay = remainCap
	}
	if pay.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	var inviter UserPO
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inviter, inviterID).Error; err != nil {
		return err
	}
	inviter.UsdtReward = inviter.UsdtReward.Add(pay)
	if err := tx.Model(&inviter).Update("usdt_reward", inviter.UsdtReward).Error; err != nil {
		return err
	}
	// accelerate inviter active orders
	left := pay
	for i := range orders {
		if left.LessThanOrEqual(decimal.Zero) {
			break
		}
		o := &orders[i]
		remain := o.ExitCap.Sub(o.EarnedTotal)
		if remain.LessThanOrEqual(decimal.Zero) {
			continue
		}
		apply := left
		if apply.GreaterThan(remain) {
			apply = remain
		}
		o.EarnedTotal = o.EarnedTotal.Add(apply)
		status := biz.OrderStatusActive
		updates := map[string]interface{}{"earned_total": o.EarnedTotal}
		if o.EarnedTotal.GreaterThanOrEqual(o.ExitCap) {
			status = biz.OrderStatusExited
			now := time.Now()
			updates["status"] = status
			updates["exited_time"] = now
		}
		if err := tx.Model(o).Updates(updates).Error; err != nil {
			return err
		}
		left = left.Sub(apply)
	}
	fromID := fromUserID
	oid := orderID
	base := directBase
	rate := decimal.NewFromFloat(directRate)
	log := &RewardLogPO{
		UserID:      inviterID,
		FromUserID:  &fromID,
		OrderID:     &oid,
		Type:        biz.RewardTypeDynamicUsdt,
		Asset:       biz.TokenUSDT,
		Amount:      pay,
		BaseAmount:  &base,
		Rate:        &rate,
		ExitApplied: pay,
	}
	return tx.Create(log).Error
}

func (r *walletRepo) ListOrdersByUser(ctx context.Context, userID int64) ([]*biz.Order, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.Order, 0, len(list))
	for i := range list {
		out = append(out, r.orderToBiz(&list[i]))
	}
	return out, nil
}

func (r *walletRepo) ListAllOrders(ctx context.Context) ([]*biz.AdminOrderDetail, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.AdminOrderDetail, 0, len(list))
	for i := range list {
		addr := ""
		var u UserPO
		if err := r.data.db.WithContext(ctx).Select("address").First(&u, list[i].UserID).Error; err == nil {
			addr = u.Address
		}
		o := r.orderToBiz(&list[i])
		o.SyncCompatFields()
		out = append(out, &biz.AdminOrderDetail{Order: o, UserAddress: addr})
	}
	return out, nil
}

func (r *walletRepo) FindOrder(ctx context.Context, id int64) (*biz.Order, error) {
	var po OrderPO
	err := r.data.db.WithContext(ctx).First(&po, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o := r.orderToBiz(&po)
	o.SyncCompatFields()
	return o, nil
}

func (r *walletRepo) RemainingExitCapacity(ctx context.Context, userID int64) (string, error) {
	var list []OrderPO
	if err := r.data.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, biz.OrderStatusActive).Find(&list).Error; err != nil {
		return "", err
	}
	total := decimal.Zero
	for _, o := range list {
		r := o.ExitCap.Sub(o.EarnedTotal)
		if r.IsPositive() {
			total = total.Add(r)
		}
	}
	return total.String(), nil
}

func (r *walletRepo) CreateTransfer(ctx context.Context, t *biz.Transfer) (*biz.Transfer, error) {
	amount, err := decimal.NewFromString(t.Amount)
	if err != nil {
		return nil, err
	}
	var created *biz.Transfer
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var from, to UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&from, t.FromUserID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&to, t.ToUserID).Error; err != nil {
			return err
		}
		po := &TransferPO{
			FromUserID: t.FromUserID,
			ToUserID:   t.ToUserID,
			Asset:      t.Asset,
			Amount:     amount,
			PayFrom:    t.PayFrom,
			Remark:     t.Remark,
		}
		switch t.Asset {
		case biz.TokenUSDT:
			if t.PayFrom != biz.PayFromReward {
				return fmt.Errorf("invalid pay_from")
			}
			if from.UsdtReward.LessThan(amount) {
				return fmt.Errorf("insufficient usdt_reward")
			}
			from.UsdtReward = from.UsdtReward.Sub(amount)
			po.FromRewardDebit = amount
			to.UsdtReward = to.UsdtReward.Add(amount)
			po.ToCreditReward = amount
		default:
			return fmt.Errorf("invalid asset")
		}
		if err := tx.Model(&from).Updates(map[string]interface{}{
			"usdt_recharge": from.UsdtRecharge,
			"usdt_reward":   from.UsdtReward,
			"aix_balance":   from.AixBalance,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&to).Updates(map[string]interface{}{
			"usdt_reward": to.UsdtReward,
			"aix_balance": to.AixBalance,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		created = r.transferToBiz(po)
		return nil
	})
	return created, err
}

func (r *walletRepo) ListTransfersByUser(ctx context.Context, userID int64) ([]*biz.Transfer, error) {
	var list []TransferPO
	if err := r.data.db.WithContext(ctx).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.Transfer, 0, len(list))
	for i := range list {
		out = append(out, r.transferToBiz(&list[i]))
	}
	return out, nil
}

func (r *walletRepo) MoveRechargeToReward(ctx context.Context, userID int64, amount string) (string, string, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return "", "", fmt.Errorf("invalid amount")
	}
	var rechargeBal, rewardBal string
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error; err != nil {
			return err
		}
		if u.UsdtRecharge.LessThan(amt) {
			return fmt.Errorf("insufficient usdt_recharge")
		}
		u.UsdtRecharge = u.UsdtRecharge.Sub(amt)
		u.UsdtReward = u.UsdtReward.Add(amt)
		if err := tx.Model(&u).Updates(map[string]interface{}{
			"usdt_recharge": u.UsdtRecharge,
			"usdt_reward":   u.UsdtReward,
		}).Error; err != nil {
			return err
		}
		po := &TransferPO{
			FromUserID:        userID,
			ToUserID:          userID,
			Asset:             biz.TokenUSDT,
			Amount:            amt,
			PayFrom:           biz.PayFromRecharge,
			FromRechargeDebit: amt,
			ToCreditReward:    amt,
			Remark:            "recharge_to_reward",
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		rechargeBal = u.UsdtRecharge.String()
		rewardBal = u.UsdtReward.String()
		return nil
	})
	return rechargeBal, rewardBal, err
}

func (r *walletRepo) CreateAixWithdrawal(ctx context.Context, userID int64, amount, toAddress string) (*biz.Withdrawal, string, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return nil, "", fmt.Errorf("invalid amount")
	}
	var created *biz.Withdrawal
	var left string
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error; err != nil {
			return err
		}
		if u.AixBalance.LessThan(amt) {
			return fmt.Errorf("insufficient aix_balance")
		}
		u.AixBalance = u.AixBalance.Sub(amt)
		if err := tx.Model(&u).Update("aix_balance", u.AixBalance).Error; err != nil {
			return err
		}
		po := &WithdrawalPO{
			UserID:    userID,
			Asset:     biz.TokenAIX,
			Amount:    amt,
			Fee:       decimal.Zero,
			PayAmount: amt,
			ToAddress: toAddress,
			Status:    biz.WithdrawStatusPending,
			Remark:    "AIX token withdraw; chain payout pending contract",
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		left = u.AixBalance.String()
		created = &biz.Withdrawal{
			ID:        po.ID,
			UserID:    po.UserID,
			Address:   u.Address,
			ToAddress: po.ToAddress,
			Amount:    po.Amount.String(),
			Fee:       po.Fee.String(),
			NetAmount: po.PayAmount.String(),
			Status:    po.Status,
			TxHash:    po.TxHash,
			CreatedAt: po.CreatedTime,
		}
		return nil
	})
	return created, left, err
}

func (r *walletRepo) CreateRewardLog(ctx context.Context, log *biz.RewardLog) error {
	amount, _ := decimal.NewFromString(log.Amount)
	po := &RewardLogPO{
		UserID:      log.UserID,
		FromUserID:  log.FromUserID,
		OrderID:     log.OrderID,
		BatchID:     log.BatchID,
		Type:        log.Type,
		Asset:       log.Asset,
		Amount:      amount,
		ExitApplied: decimal.Zero,
	}
	if log.BaseAmount != "" {
		b, _ := decimal.NewFromString(log.BaseAmount)
		po.BaseAmount = &b
	}
	if log.Rate != "" {
		rt, _ := decimal.NewFromString(log.Rate)
		po.Rate = &rt
	}
	if log.ExitApplied != "" {
		ea, _ := decimal.NewFromString(log.ExitApplied)
		po.ExitApplied = ea
	}
	if log.SettlementDate != "" {
		d := log.SettlementDate
		po.SettlementDate = &d
	}
	if log.Meta != "" {
		m := log.Meta
		po.Meta = &m
	}
	return r.data.db.WithContext(ctx).Create(po).Error
}

func (r *walletRepo) ListRewardLogsByUser(ctx context.Context, userID int64) ([]*biz.RewardLog, error) {
	var list []RewardLogPO
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.RewardLog, 0, len(list))
	for i := range list {
		out = append(out, rewardLogToBiz(&list[i]))
	}
	return out, nil
}

func (r *walletRepo) GetAixPrice(ctx context.Context, date string) (string, error) {
	var po AixPricePO
	err := r.data.db.WithContext(ctx).Where("effective_date = ?", date).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return po.Price.String(), nil
}

func (r *walletRepo) UpsertAixPrice(ctx context.Context, date, price, remark string) error {
	p, err := decimal.NewFromString(price)
	if err != nil {
		return err
	}
	var po AixPricePO
	err = r.data.db.WithContext(ctx).Where("effective_date = ?", date).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return r.data.db.WithContext(ctx).Create(&AixPricePO{Price: p, EffectiveDate: date, Remark: remark}).Error
	}
	if err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&po).Updates(map[string]interface{}{"price": p, "remark": remark}).Error
}

func (r *walletRepo) rechargeToBiz(po *RechargePO) *biz.Recharge {
	rec := &biz.Recharge{
		ID:          po.ID,
		UserID:      po.UserID,
		Address:     po.FromAddress,
		Amount:      po.Amount.String(),
		Message:     po.Message,
		TxHash:      po.TxHash,
		FromAddress: po.FromAddress,
		ToAddress:   po.ToAddress,
		Status:      po.Status,
		CreatedAt:   po.CreatedTime,
		CreatedTime: po.CreatedTime,
	}
	if po.ExpireAt != nil {
		rec.ExpireAt = *po.ExpireAt
	}
	if po.ConfirmedTime != nil {
		rec.ConfirmedAt = po.ConfirmedTime
		rec.ConfirmedTime = po.ConfirmedTime
	}
	return rec
}

func (r *walletRepo) orderToBiz(po *OrderPO) *biz.Order {
	o := &biz.Order{
		ID:           po.ID,
		UserID:       po.UserID,
		Principal:    po.Principal.String(),
		ExitCap:      po.ExitCap.String(),
		EarnedTotal:  po.EarnedTotal.String(),
		DirectBase:   po.DirectBase.String(),
		FromRecharge: po.FromRecharge.String(),
		FromReward:   po.FromReward.String(),
		FundSource:   po.FundSource,
		Status:       po.Status,
		ExitedTime:   po.ExitedTime,
		CreatedTime:  po.CreatedTime,
		UpdatedTime:  po.UpdatedTime,
	}
	o.SyncCompatFields()
	return o
}

func (r *walletRepo) transferToBiz(po *TransferPO) *biz.Transfer {
	return &biz.Transfer{
		ID:                po.ID,
		FromUserID:        po.FromUserID,
		ToUserID:          po.ToUserID,
		Asset:             po.Asset,
		Amount:            po.Amount.String(),
		PayFrom:           po.PayFrom,
		FromRechargeDebit: po.FromRechargeDebit.String(),
		FromRewardDebit:   po.FromRewardDebit.String(),
		ToCreditReward:    po.ToCreditReward.String(),
		ToCreditAix:       po.ToCreditAix.String(),
		Remark:            po.Remark,
		CreatedTime:       po.CreatedTime,
	}
}

func rewardLogToBiz(po *RewardLogPO) *biz.RewardLog {
	log := &biz.RewardLog{
		ID:          po.ID,
		UserID:      po.UserID,
		FromUserID:  po.FromUserID,
		OrderID:     po.OrderID,
		BatchID:     po.BatchID,
		Type:        po.Type,
		Asset:       po.Asset,
		Amount:      po.Amount.String(),
		ExitApplied: po.ExitApplied.String(),
		CreatedTime: po.CreatedTime,
	}
	if po.BaseAmount != nil {
		log.BaseAmount = po.BaseAmount.String()
	}
	if po.Rate != nil {
		log.Rate = po.Rate.String()
	}
	if po.SettlementDate != nil {
		log.SettlementDate = *po.SettlementDate
	}
	if po.Meta != nil {
		log.Meta = *po.Meta
	}
	return log
}

// --- legacy stubs ---

func (r *walletRepo) CreateWithdrawal(ctx context.Context, userID int64, amount, fee, netAmount, toAddress string) (*biz.Withdrawal, string, error) {
	return nil, "", fmt.Errorf("USDT withdraw not supported; use AIX withdraw")
}
func (r *walletRepo) ListWithdrawalsByUser(ctx context.Context, userID int64) ([]*biz.Withdrawal, error) {
	var list []WithdrawalPO
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.Withdrawal, 0, len(list))
	for _, po := range list {
		out = append(out, &biz.Withdrawal{
			ID: po.ID, UserID: po.UserID, ToAddress: po.ToAddress,
			Amount: po.Amount.String(), Fee: po.Fee.String(), NetAmount: po.PayAmount.String(),
			Status: po.Status, TxHash: po.TxHash, CreatedAt: po.CreatedTime,
		})
	}
	return out, nil
}
func (r *walletRepo) CreateClaimRecord(ctx context.Context, record *biz.ClaimRecord) error {
	return fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) ListClaimRecordsByUser(ctx context.Context, userID int64) ([]*biz.ClaimRecord, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) SumClaimedByUser(ctx context.Context, userID int64) (string, error) {
	return "0", nil
}
func (r *walletRepo) SumWithdrawnByUser(ctx context.Context, userID int64) (string, error) {
	return "0", nil
}
func (r *walletRepo) ListAllWithdrawals(ctx context.Context) ([]*biz.Withdrawal, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) ApproveWithdrawal(ctx context.Context, id int64) error {
	return fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) ListProducts(ctx context.Context) ([]*biz.Product, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) FindProduct(ctx context.Context, id int64) (*biz.Product, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) FindProductByPrice(ctx context.Context, price string) (*biz.Product, error) {
	return nil, nil
}
func (r *walletRepo) SubscribeProduct(ctx context.Context, userID, productID int64, quantity int32, totalAmount string) (*biz.Order, string, error) {
	return nil, "", fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) SubscribeByAmount(ctx context.Context, userID int64, totalAmount, productName string) (*biz.Order, string, error) {
	return nil, "", fmt.Errorf("not supported in AIX: use Subscribe with pay_from")
}
func (r *walletRepo) ListAllProducts(ctx context.Context) ([]*biz.Product, error) {
	return []*biz.Product{}, nil
}
func (r *walletRepo) CreateProduct(ctx context.Context, product *biz.Product) (*biz.Product, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) AdminUpdateProduct(ctx context.Context, product *biz.Product) (*biz.Product, error) {
	return nil, fmt.Errorf("not supported in AIX")
}
func (r *walletRepo) AdminUpdateOrder(ctx context.Context, update *biz.AdminOrderUpdate) (*biz.Order, error) {
	updates := map[string]interface{}{}
	if update.TotalAmount != "" {
		v, err := decimal.NewFromString(update.TotalAmount)
		if err != nil {
			return nil, err
		}
		updates["principal"] = v
	}
	if update.ExitTarget != "" {
		v, err := decimal.NewFromString(update.ExitTarget)
		if err != nil {
			return nil, err
		}
		updates["exit_cap"] = v
	}
	if update.ReleasedAmount != "" {
		v, err := decimal.NewFromString(update.ReleasedAmount)
		if err != nil {
			return nil, err
		}
		updates["earned_total"] = v
	}
	if update.Status != "" {
		st := update.Status
		if st == "completed" {
			st = biz.OrderStatusExited
		}
		updates["status"] = st
	}
	if len(updates) == 0 {
		return r.FindOrder(ctx, update.OrderID)
	}
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order OrderPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, update.OrderID).Error; err != nil {
			return err
		}
		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}
		return refreshAncestorPerformance(tx, order.UserID)
	})
	if err != nil {
		return nil, err
	}
	return r.FindOrder(ctx, update.OrderID)
}
