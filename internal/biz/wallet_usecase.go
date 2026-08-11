package biz

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"backend/internal/conf"
	"backend/internal/pkg/eth"
	"backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

var errAIXUnsupported = errors.BadRequest("NOT_SUPPORTED", "not supported in AIX")

// WalletUsecase handles recharge, subscribe and transfer logic.
type WalletUsecase struct {
	userRepo    UserRepo
	walletRepo  WalletRepo
	stakingRepo StakingRepo
	authCfg     *conf.AuthConfig
	walletCfg   *conf.WalletConfig
	log         *log.Helper
}

func NewWalletUsecase(userRepo UserRepo, walletRepo WalletRepo, stakingRepo StakingRepo, authCfg *conf.AuthConfig, walletCfg *conf.WalletConfig, logger log.Logger) *WalletUsecase {
	return &WalletUsecase{
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		stakingRepo: stakingRepo,
		authCfg:     authCfg,
		walletCfg:   walletCfg,
		log:         log.NewHelper(logger),
	}
}

func (uc *WalletUsecase) resolveUser(ctx context.Context, tokenString string) (*User, error) {
	address, err := token.Parse(tokenString, uc.authCfg.GetJwtSecret())
	if err != nil {
		return nil, errors.Unauthorized("UNAUTHORIZED", "token 无效或已过期")
	}
	user, err := uc.userRepo.FindByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("USER_NOT_FOUND", "用户不存在")
	}
	return user, nil
}

// GetBalance returns user + AIX balances mapped into legacy string slots.
// balance=usdt_recharge, released=usdt_reward, claimed=aix_balance, pending=daily static estimate, unexited=remaining exit cap
func (uc *WalletUsecase) GetBalance(ctx context.Context, tokenString string) (*User, *OrderReleaseSummary, string, string, string, string, string, int64, int64, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, nil, "", "", "", "", "", 0, 0, err
	}
	orders, err := uc.walletRepo.ListOrdersByUser(ctx, user.ID)
	if err != nil {
		return nil, nil, "", "", "", "", "", 0, 0, err
	}
	summary := SummarizeOrders(orders)
	pending := summary.PendingTotal
	unexited := summary.UnexitedTotal
	now := token.NowChina()
	nextRelease := token.NextChinaMidnight(now)
	return user, &summary, user.UsdtRecharge, user.UsdtReward, pending, user.AixBalance, unexited, nextRelease.Unix(), now.Unix(), nil
}

func SummarizeOrders(orders []*Order) OrderReleaseSummary {
	exitTotal := decimal.Zero
	earnedTotal := decimal.Zero
	pending := decimal.Zero
	unexited := decimal.Zero
	rate := decimal.NewFromFloat(StaticRate).Div(decimal.NewFromInt(100))
	nodes := int32(0)
	for _, o := range orders {
		nodes++
		p, _ := decimal.NewFromString(o.Principal)
		cap, _ := decimal.NewFromString(o.ExitCap)
		earned, _ := decimal.NewFromString(o.EarnedTotal)
		exitTotal = exitTotal.Add(cap)
		earnedTotal = earnedTotal.Add(earned)
		remain := cap.Sub(earned)
		if remain.IsPositive() {
			unexited = unexited.Add(remain)
		}
		if o.Status == OrderStatusActive {
			day := p.Mul(rate)
			if day.GreaterThan(remain) && remain.IsPositive() {
				day = remain
			}
			if remain.IsPositive() {
				pending = pending.Add(day)
			}
		}
	}
	return OrderReleaseSummary{
		ExitTotal:     exitTotal.String(),
		ReleasedTotal: earnedTotal.String(),
		PendingTotal:  pending.String(),
		UnexitedTotal: unexited.String(),
		TotalNodes:    nodes,
	}
}

func (uc *WalletUsecase) IsDevMode() bool {
	return uc.walletCfg.GetRPCURL() == ""
}

func (uc *WalletUsecase) CreateRecharge(ctx context.Context, tokenString, amount string) (*Recharge, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	amountDec, err := ParseAmount(amount)
	minRecharge := decimal.NewFromInt(5)
	if err != nil || amountDec.LessThan(minRecharge) {
		return nil, errors.BadRequest("INVALID_AMOUNT", "USDT 充值金额不能小于5")
	}
	depositAddress := uc.walletCfg.GetDepositAddress()
	if !uc.IsDevMode() {
		if depositAddress == "" || depositAddress == ZeroAddress {
			return nil, errors.BadRequest("DEPOSIT_NOT_CONFIGURED", "平台 USDT 收款地址未配置，请联系管理员")
		}
	}
	if uc.walletCfg.GetUsdtContract() == "" {
		return nil, errors.BadRequest("USDT_NOT_CONFIGURED", "USDT 合约地址未配置")
	}
	now := time.Now()
	message := fmt.Sprintf(
		"Recharge USDT to AIX account\nAddress: %s\nAmount: %s USDT\nToken: %s\nRechargeAt: %d",
		user.Address, amountDec.String(), uc.walletCfg.GetUsdtContract(), now.Unix(),
	)
	recharge := &Recharge{
		UserID:      user.ID,
		Address:     user.Address,
		FromAddress: user.Address,
		ToAddress:   depositAddress,
		Amount:      amountDec.String(),
		Message:     message,
		Status:      RechargeStatusPending,
		ExpireAt:    now.Add(30 * time.Minute),
	}
	return uc.walletRepo.CreateRecharge(ctx, recharge)
}

func (uc *WalletUsecase) ConfirmRecharge(ctx context.Context, tokenString string, rechargeID int64, txHash string, txHashes []string, signature string) (string, string, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return "", "", err
	}
	recharge, err := uc.walletRepo.FindRecharge(ctx, rechargeID)
	if err != nil {
		return "", "", err
	}
	if recharge == nil {
		return "", "", errors.NotFound("RECHARGE_NOT_FOUND", "充值记录不存在")
	}
	if recharge.UserID != user.ID {
		return "", "", errors.Forbidden("RECHARGE_FORBIDDEN", "无权确认该充值记录")
	}
	if recharge.Status == RechargeStatusConfirmed {
		return "", "", errors.BadRequest("RECHARGE_CONFIRMED", "充值记录已确认")
	}
	if !recharge.ExpireAt.IsZero() && time.Now().After(recharge.ExpireAt) {
		return "", "", errors.BadRequest("RECHARGE_EXPIRED", "充值单已过期，请重新创建")
	}
	hashes := normalizeTxHashes(txHash, txHashes)
	if len(hashes) == 0 {
		return "", "", errors.BadRequest("INVALID_TX_HASH", "交易哈希不能为空")
	}
	for _, h := range hashes {
		exists, err := uc.walletRepo.FindRechargeByTxHash(ctx, h)
		if err != nil {
			return "", "", err
		}
		if exists != nil && exists.ID != rechargeID {
			return "", "", errors.BadRequest("TX_HASH_USED", "交易哈希已被使用")
		}
	}
	if err := eth.VerifyPersonalSign(recharge.Message, signature, user.Address); err != nil {
		return "", "", errors.Unauthorized("INVALID_SIGNATURE", "签名校验失败")
	}
	amountDec, _ := ParseAmount(recharge.Amount)
	depositAddrs := uc.walletCfg.GetDepositAddresses()
	splits := SplitEqualAmounts(amountDec, len(depositAddrs), uc.walletCfg.GetUsdtDecimals())
	joinedHash := strings.Join(hashes, ",")

	if uc.walletCfg.GetRPCURL() != "" {
		if len(depositAddrs) == 0 {
			return "", "", errors.BadRequest("DEPOSIT_NOT_CONFIGURED", "平台 USDT 收款地址未配置，请联系管理员")
		}
		if len(depositAddrs) == 1 {
			if len(hashes) != 1 {
				return "", "", errors.BadRequest("INVALID_TX_HASH", "请提交 1 笔充值交易哈希")
			}
			if err := eth.VerifyUSDTTransfer(ctx, uc.walletCfg.GetRPCURL(), hashes[0], uc.walletCfg.GetUsdtContract(), depositAddrs, user.Address, amountDec, uc.walletCfg.GetUsdtDecimals()); err != nil {
				return "", "", errors.BadRequest("TX_VERIFY_FAILED", err.Error())
			}
		} else {
			if len(hashes) != len(depositAddrs) {
				return "", "", errors.BadRequest("INVALID_TX_HASH", fmt.Sprintf("请分别向 %d 个收款地址各转账一笔，并提交全部交易哈希", len(depositAddrs)))
			}
			used := make([]bool, len(hashes))
			for i, addr := range depositAddrs {
				matched := false
				var lastErr error
				for j, h := range hashes {
					if used[j] {
						continue
					}
					err := eth.VerifyUSDTTransfer(ctx, uc.walletCfg.GetRPCURL(), h, uc.walletCfg.GetUsdtContract(), []string{addr}, user.Address, splits[i], uc.walletCfg.GetUsdtDecimals())
					if err == nil {
						used[j] = true
						matched = true
						break
					}
					lastErr = err
				}
				if !matched {
					msg := "未找到对应收款地址的分账转账"
					if lastErr != nil {
						msg = lastErr.Error()
					}
					return "", "", errors.BadRequest("TX_VERIFY_FAILED", fmt.Sprintf("收款地址 %s 校验失败: %s", addr, msg))
				}
			}
		}
	}

	balance, err := uc.walletRepo.ConfirmRechargeCredit(ctx, rechargeID, joinedHash)
	if err != nil {
		return "", "", err
	}
	return balance, recharge.Amount, nil
}

func normalizeTxHashes(txHash string, txHashes []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(txHashes)+1)
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '|' || r == ' ' || r == '\n' || r == '\t'
		})
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			key := strings.ToLower(p)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, p)
		}
	}
	for _, h := range txHashes {
		add(h)
	}
	add(txHash)
	return out
}

func SplitEqualAmounts(total decimal.Decimal, n int, decimals int32) []decimal.Decimal {
	if n <= 0 {
		return nil
	}
	if decimals < 0 {
		decimals = 0
	}
	if n == 1 {
		return []decimal.Decimal{total}
	}
	out := make([]decimal.Decimal, n)
	base := total.Div(decimal.NewFromInt(int64(n))).Truncate(decimals)
	assigned := decimal.Zero
	for i := 0; i < n-1; i++ {
		out[i] = base
		assigned = assigned.Add(base)
	}
	out[n-1] = total.Sub(assigned)
	return out
}

func (uc *WalletUsecase) ListRecharges(ctx context.Context, tokenString string) ([]*Recharge, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return uc.walletRepo.ListRechargesByUser(ctx, user.ID)
}

func (uc *WalletUsecase) ClaimToAccount(ctx context.Context, tokenString, amount string) (string, string, string, error) {
	return "", "", "", errAIXUnsupported
}

func (uc *WalletUsecase) ListClaimRecords(ctx context.Context, tokenString string) ([]*ClaimRecord, error) {
	return nil, errAIXUnsupported
}

func (uc *WalletUsecase) CreateWithdraw(ctx context.Context, tokenString, amount, toAddress, signature string, withdrawAt int64) (*Withdrawal, string, error) {
	return nil, "", errors.BadRequest("USDT_WITHDRAW_FORBIDDEN", "仅支持提现 AIX 代币，不支持提现 USDT")
}

// CreateAixWithdraw 提现 AIX 代币（合约未配置时仅扣账并记 pending，链上打款后续补齐）
func (uc *WalletUsecase) CreateAixWithdraw(ctx context.Context, tokenString, amount, toAddress string) (*Withdrawal, string, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, "", err
	}
	amt, err := ParseAmount(amount)
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return nil, "", errors.BadRequest("INVALID_AMOUNT", "提现金额必须大于0")
	}
	toNorm := strings.TrimSpace(toAddress)
	if toNorm == "" {
		toNorm = user.Address
	} else {
		toNorm, err = eth.NormalizeAddress(toNorm)
		if err != nil {
			return nil, "", errors.BadRequest("INVALID_ADDRESS", "提现地址无效")
		}
	}
	// AIX 合约地址后续配置；此处不校验链上合约
	w, left, err := uc.walletRepo.CreateAixWithdrawal(ctx, user.ID, amt.String(), toNorm)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return nil, "", errors.BadRequest("INSUFFICIENT_AIX", "AIX 代币余额不足")
		}
		return nil, "", err
	}
	return w, left, nil
}

func (uc *WalletUsecase) ListWithdrawals(ctx context.Context, tokenString string) ([]*Withdrawal, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return uc.walletRepo.ListWithdrawalsByUser(ctx, user.ID)
}

// MoveRechargeToReward 充值钱包 USDT → 奖励钱包（同账户，不产生直推）
func (uc *WalletUsecase) MoveRechargeToReward(ctx context.Context, tokenString, amount string) (string, string, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return "", "", err
	}
	amt, err := ParseAmount(amount)
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return "", "", errors.BadRequest("INVALID_AMOUNT", "划转金额必须大于0")
	}
	rechargeBal, rewardBal, err := uc.walletRepo.MoveRechargeToReward(ctx, user.ID, amt.String())
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return "", "", errors.BadRequest("INSUFFICIENT_BALANCE", "充值钱包余额不足")
		}
		return "", "", err
	}
	return rechargeBal, rewardBal, nil
}

func (uc *WalletUsecase) DepositAddress() string {
	return uc.walletCfg.GetDepositAddress()
}

func (uc *WalletUsecase) DepositAddresses() []string {
	return uc.walletCfg.GetDepositAddresses()
}

func (uc *WalletUsecase) SplitDepositAmounts(amount string) []string {
	total, err := ParseAmount(amount)
	if err != nil || !total.GreaterThan(decimal.Zero) {
		return nil
	}
	addrs := uc.walletCfg.GetDepositAddresses()
	parts := SplitEqualAmounts(total, len(addrs), uc.walletCfg.GetUsdtDecimals())
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, p.String())
	}
	return out
}

func (uc *WalletUsecase) UsdtContract() string {
	return uc.walletCfg.GetUsdtContract()
}

func (uc *WalletUsecase) ListProducts(ctx context.Context) ([]*Product, error) {
	return nil, errAIXUnsupported
}

// Subscribe AIX 报单：amount + pay_from(recharge|reward)
func (uc *WalletUsecase) Subscribe(ctx context.Context, tokenString string, productID int64, quantity int32, amountStr string) (*Order, string, error) {
	// Legacy proto path without pay_from — reject and ask for custom route
	return nil, "", errors.BadRequest("PAY_FROM_REQUIRED", "请使用 /v1/wallet/subscribe 并传 pay_from=recharge|reward")
}

// SubscribeAIX 报单 / 复投
func (uc *WalletUsecase) SubscribeAIX(ctx context.Context, tokenString, amountStr, payFrom string) (*Order, string, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, "", err
	}
	payFrom = strings.ToLower(strings.TrimSpace(payFrom))
	if payFrom != PayFromRecharge && payFrom != PayFromReward {
		return nil, "", errors.BadRequest("INVALID_PAY_FROM", "pay_from 必须为 recharge 或 reward")
	}
	minSubscribe, err := ParseAmount(uc.walletCfg.GetMinSubscribe())
	if err != nil {
		minSubscribe = decimal.NewFromInt(100)
	}
	total, err := ParseAmount(strings.TrimSpace(amountStr))
	if err != nil || !total.GreaterThan(decimal.Zero) {
		return nil, "", errors.BadRequest("INVALID_AMOUNT", "认购金额必须大于0")
	}
	if total.LessThan(minSubscribe) {
		return nil, "", errors.BadRequest("MIN_SUBSCRIBE_LIMIT", fmt.Sprintf("认购金额不能低于 %s USDT", minSubscribe.String()))
	}
	order, bal, err := uc.walletRepo.Subscribe(ctx, user.ID, total.String(), payFrom, ExitMultiplier, DirectRate)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return nil, "", errors.BadRequest("INSUFFICIENT_BALANCE", "账户余额不足")
		}
		return nil, "", err
	}
	order.SyncCompatFields()
	return order, bal, nil
}

func (uc *WalletUsecase) ListOrders(ctx context.Context, tokenString string) ([]*Order, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	orders, err := uc.walletRepo.ListOrdersByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		o.SyncCompatFields()
	}
	return orders, nil
}

// Transfer 上下级转账
func (uc *WalletUsecase) Transfer(ctx context.Context, tokenString, toAddress, asset, amount, payFrom string) (*Transfer, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	toNorm, err := eth.NormalizeAddress(toAddress)
	if err != nil {
		return nil, errors.BadRequest("INVALID_ADDRESS", "收款地址无效")
	}
	toUser, err := uc.userRepo.FindByAddress(ctx, toNorm)
	if err != nil {
		return nil, err
	}
	if toUser == nil {
		return nil, errors.NotFound("USER_NOT_FOUND", "收款用户不存在")
	}
	if toUser.ID == user.ID {
		return nil, errors.BadRequest("INVALID_TRANSFER", "不能转给自己")
	}
	ok, err := uc.userRepo.IsUplineOrDownline(ctx, user.ID, toUser.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.BadRequest("NOT_UPLINE_DOWNLINE", "仅允许邀请树上下级互转")
	}
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = TokenUSDT
	}
	if asset != TokenUSDT {
		return nil, errors.BadRequest("INVALID_ASSET", "上下级互转仅支持奖励钱包 USDT")
	}
	amt, err := ParseAmount(amount)
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return nil, errors.BadRequest("INVALID_AMOUNT", "转账金额必须大于0")
	}
	payFrom = strings.ToLower(strings.TrimSpace(payFrom))
	if payFrom != PayFromReward {
		return nil, errors.BadRequest("INVALID_PAY_FROM", "上下级互转只能从奖励钱包扣款")
	}
	t := &Transfer{
		FromUserID: user.ID,
		ToUserID:   toUser.ID,
		Asset:      asset,
		Amount:     amt.String(),
		PayFrom:    payFrom,
	}
	created, err := uc.walletRepo.CreateTransfer(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return nil, errors.BadRequest("INSUFFICIENT_BALANCE", "余额不足")
		}
		return nil, err
	}
	return created, nil
}

func (uc *WalletUsecase) ListRewardLogs(ctx context.Context, tokenString string) ([]*RewardLog, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return uc.walletRepo.ListRewardLogsByUser(ctx, user.ID)
}

func (uc *WalletUsecase) GetAixPrice(ctx context.Context, date string) (string, error) {
	if date == "" {
		date = token.NowChina().Format("2006-01-02")
	}
	price, err := uc.walletRepo.GetAixPrice(ctx, date)
	if err != nil {
		return "", err
	}
	if price == "" {
		return decimal.NewFromFloat(AixPriceInitial).String(), nil
	}
	return price, nil
}

func (uc *WalletUsecase) ListReleaseRecords(ctx context.Context, tokenString string) ([]*ReleaseRecord, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return uc.stakingRepo.ListReleaseRecordsByUser(ctx, user.ID)
}

func (uc *WalletUsecase) ListReferralRewards(ctx context.Context, tokenString string) ([]*ReferralReward, error) {
	user, err := uc.resolveUser(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return uc.stakingRepo.ListReferralRewardsByUser(ctx, user.ID)
}

func (uc *WalletUsecase) SumReferralByOrderDate(ctx context.Context, orderID int64, settlementDate string) (string, error) {
	return uc.stakingRepo.SumReferralByOrderDate(ctx, orderID, settlementDate)
}

func (uc *WalletUsecase) FindOrder(ctx context.Context, orderID int64) (*Order, error) {
	return uc.walletRepo.FindOrder(ctx, orderID)
}

func (uc *WalletUsecase) FindUserAddress(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", err
	}
	return user.Address, nil
}

func (uc *WalletUsecase) BuildOrderIndexMap(ctx context.Context, userID int64) (map[int64]int32, error) {
	orders, err := uc.walletRepo.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	sorted := append([]*Order(nil), orders...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	result := make(map[int64]int32, len(sorted))
	for i, o := range sorted {
		result[o.ID] = int32(i + 1)
	}
	return result, nil
}

func (uc *WalletUsecase) ListEcoRewards(ctx context.Context, tokenString string) ([]*EcoReward, error) {
	return nil, errAIXUnsupported
}
