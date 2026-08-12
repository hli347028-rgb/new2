package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/internal/conf"
	"backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	ZeroAddress = "0x0000000000000000000000000000000000000000"
	RoleAdmin   = "admin"
	RoleUser    = "user"
)

type AdminUserDetail struct {
	User          *User
	InviteeCount  int32
	WithdrawReset bool
}

type SettingsRepo interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
}

type AdminUsecase struct {
	userRepo     UserRepo
	walletRepo   WalletRepo
	settingsRepo SettingsRepo
	settlement   *SettlementUsecase
	authCfg      *conf.AuthConfig
	walletCfg    *conf.WalletConfig
	log          *log.Helper
}

func NewAdminUsecase(
	userRepo UserRepo,
	walletRepo WalletRepo,
	settingsRepo SettingsRepo,
	settlement *SettlementUsecase,
	authCfg *conf.AuthConfig,
	walletCfg *conf.WalletConfig,
	logger log.Logger,
) *AdminUsecase {
	return &AdminUsecase{
		userRepo:     userRepo,
		walletRepo:   walletRepo,
		settingsRepo: settingsRepo,
		settlement:   settlement,
		authCfg:      authCfg,
		walletCfg:    walletCfg,
		log:          log.NewHelper(logger),
	}
}

func IsAdmin(user *User, authCfg *conf.AuthConfig) bool {
	if user == nil {
		return false
	}
	if user.Role == RoleAdmin || user.Address == ZeroAddress {
		return true
	}
	for _, addr := range authCfg.GetAdminAddresses() {
		if strings.EqualFold(addr, user.Address) {
			return true
		}
	}
	return false
}

func (uc *AdminUsecase) requireAdmin(ctx context.Context, tokenString string) (*User, error) {
	address, err := token.Parse(tokenString, uc.authCfg.GetJwtSecret())
	if err != nil {
		return nil, errors.Unauthorized("UNAUTHORIZED", "token 无效或已过期")
	}
	user, err := uc.userRepo.FindByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	if user == nil || !IsAdmin(user, uc.authCfg) {
		return nil, errors.Forbidden("FORBIDDEN", "需要管理员权限")
	}
	return user, nil
}

func (uc *AdminUsecase) RequireAdminUser(ctx context.Context, tokenString string) (*User, error) {
	return uc.requireAdmin(ctx, tokenString)
}

func (uc *AdminUsecase) GetPersistedConfigSnapshot() *conf.SystemConfigSnapshot {
	return uc.buildConfigSnapshot()
}

func (uc *AdminUsecase) ListUsers(ctx context.Context, tokenString string) ([]*AdminUserDetail, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	if err := uc.userRepo.RefreshPerformance(ctx); err != nil {
		return nil, err
	}
	users, err := uc.userRepo.ListAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*AdminUserDetail, 0, len(users))
	for _, user := range users {
		count, _ := uc.userRepo.CountInvitees(ctx, user.ID)
		result = append(result, &AdminUserDetail{User: user, InviteeCount: count})
	}
	return result, nil
}

type AdminUserUpdate struct {
	UserID            int64
	Balance           string
	ReleasedBalance   string
	UsdtRecharge      string
	UsdtReward        string
	AixBalance        string
	WinBalance        string
	PendingMgmtReward string
	StaticUsdtTotal   string
	Role              string
	CommunityLevel    string
	SetCommunityLevel bool
	CommunityStake    string
	TeamStake         string
	InviterID         *int64
	WithdrawReset     *bool
}

func (uc *AdminUsecase) UpdateUser(ctx context.Context, tokenString string, update *AdminUserUpdate) (*AdminUserDetail, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	if err := uc.userRepo.AdminUpdateUser(ctx, update); err != nil {
		return nil, err
	}
	user, err := uc.userRepo.FindByID(ctx, update.UserID)
	if err != nil {
		return nil, err
	}
	count, _ := uc.userRepo.CountInvitees(ctx, user.ID)
	return &AdminUserDetail{User: user, InviteeCount: count}, nil
}

func (uc *AdminUsecase) GetSystemConfig(ctx context.Context, tokenString string) (*conf.SystemConfigSnapshot, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	return uc.buildConfigSnapshot(), nil
}

func (uc *AdminUsecase) UpdateSystemConfig(ctx context.Context, tokenString string, snapshot *conf.SystemConfigSnapshot) (*conf.SystemConfigSnapshot, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	uc.applyConfigSnapshot(snapshot)
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	if err := uc.settingsRepo.Set(ctx, conf.SettingsKeySystemConfig, string(data)); err != nil {
		return nil, err
	}
	// 后台改 WIN 价时，同步覆盖 win_prices 唯一一行
	if snapshot != nil && snapshot.WinPrice > 0 {
		_ = uc.walletRepo.UpsertCurrentWinPrice(ctx, strconv.FormatFloat(snapshot.WinPrice, 'f', -1, 64), "admin")
	}
	return uc.buildConfigSnapshot(), nil
}

func (uc *AdminUsecase) buildConfigSnapshot() *conf.SystemConfigSnapshot {
	addrs := uc.walletCfg.GetDepositAddresses()
	primary := ""
	if len(addrs) > 0 {
		primary = addrs[0]
	}
	snap := &conf.SystemConfigSnapshot{
		JwtSecret:            uc.authCfg.JwtSecret,
		ChallengeTTL:         uc.authCfg.ChallengeTTL,
		AdminAddresses:       uc.authCfg.AdminAddresses,
		DepositAddress:       primary,
		DepositAddresses:     addrs,
		UsdtContract:         uc.walletCfg.UsdtContract,
		UsdtDecimals:         uc.walletCfg.UsdtDecimals,
		RPCURL:               uc.walletCfg.RPCURL,
		MinSubscribe:         uc.walletCfg.MinSubscribe,
		StaticRate:           StaticRate,
		ExitMultiplier:       ExitMultiplier,
		DirectRate:           DirectRate,
		MgmtThresholds:       append([]float64(nil), MgmtThresholds...),
		MgmtRates:            append([]float64(nil), MgmtRates...),
		AixPriceInitial:      AixPriceInitial,
		WinPrice:             WinPrice,
		ExchangeFeeRate:      ExchangeFeeRate,
		MgmtCountsTowardExit: MgmtCountsTowardExit,
	}
	conf.NormalizeBusinessDefaults(snap)
	return snap
}

func (uc *AdminUsecase) applyConfigSnapshot(snapshot *conf.SystemConfigSnapshot) {
	if snapshot == nil {
		return
	}
	conf.NormalizeBusinessDefaults(snapshot)
	if snapshot.JwtSecret != "" {
		uc.authCfg.JwtSecret = snapshot.JwtSecret
	}
	if snapshot.ChallengeTTL != "" {
		uc.authCfg.ChallengeTTL = snapshot.ChallengeTTL
	}
	if snapshot.AdminAddresses != nil {
		uc.authCfg.AdminAddresses = snapshot.AdminAddresses
	}
	if len(snapshot.DepositAddresses) > 0 {
		uc.walletCfg.SetDepositAddresses(snapshot.DepositAddresses)
	} else if snapshot.DepositAddress != "" {
		uc.walletCfg.SetDepositAddresses([]string{snapshot.DepositAddress})
	}
	if snapshot.UsdtContract != "" {
		uc.walletCfg.UsdtContract = snapshot.UsdtContract
	}
	if snapshot.UsdtDecimals > 0 {
		uc.walletCfg.UsdtDecimals = snapshot.UsdtDecimals
	}
	uc.walletCfg.RPCURL = snapshot.RPCURL
	if snapshot.MinSubscribe != "" {
		uc.walletCfg.MinSubscribe = snapshot.MinSubscribe
	}
	ApplyAixConfig(snapshot)
}

func (uc *AdminUsecase) LoadPersistedConfig(ctx context.Context) error {
	raw, err := uc.settingsRepo.Get(ctx, conf.SettingsKeySystemConfig)
	if err != nil || raw == "" {
		// 仍尝试加载唯一 WIN 现价
	} else {
		var snapshot conf.SystemConfigSnapshot
		if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
			return err
		}
		uc.applyConfigSnapshot(&snapshot)
	}
	if priceStr, err := uc.walletRepo.GetCurrentWinPrice(ctx); err == nil && strings.TrimSpace(priceStr) != "" {
		if p, err := decimal.NewFromString(strings.TrimSpace(priceStr)); err == nil && p.IsPositive() {
			f, _ := p.Float64()
			WinPrice = f
		}
	}
	// 将配置中的 AIX 现价同步到当日 aix_prices，避免种子价 1 覆盖后台配置
	if AixPriceInitial > 0 {
		date := token.NowChina().Format("2006-01-02")
		_ = uc.walletRepo.UpsertAixPrice(ctx, date, strconv.FormatFloat(AixPriceInitial, 'f', -1, 64), "sync from config")
	}
	return nil
}

// SetWinPriceFromOracle 由链上预言机写入 WIN 价格（内存 + win_prices 唯一一行）。
func (uc *AdminUsecase) SetWinPriceFromOracle(ctx context.Context, price float64) error {
	if price <= 0 {
		return fmt.Errorf("invalid win price: %v", price)
	}
	priceStr := strconv.FormatFloat(price, 'f', -1, 64)
	if err := uc.walletRepo.UpsertCurrentWinPrice(ctx, priceStr, "oracle"); err != nil {
		return err
	}
	WinPrice = price
	return nil
}

func (uc *AdminUsecase) ListAllProducts(ctx context.Context, tokenString string) ([]*Product, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	return nil, errors.BadRequest("NOT_SUPPORTED", "not supported in AIX")
}

func (uc *AdminUsecase) UpdateProduct(ctx context.Context, tokenString string, product *Product) (*Product, error) {
	return nil, errors.BadRequest("NOT_SUPPORTED", "not supported in AIX")
}

func (uc *AdminUsecase) CreateProduct(ctx context.Context, tokenString string, product *Product) (*Product, error) {
	return nil, errors.BadRequest("NOT_SUPPORTED", "not supported in AIX")
}

type AdminDownlineNode struct {
	User            *User
	DirectCount     int32
	RecommendAmount string
}

func (uc *AdminUsecase) ListUserDownline(ctx context.Context, tokenString string, userID int64, address string) ([]*AdminDownlineNode, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	var user *User
	var err error
	if userID > 0 {
		user, err = uc.userRepo.FindByID(ctx, userID)
	} else if strings.TrimSpace(address) != "" {
		user, err = uc.userRepo.FindByAddress(ctx, strings.TrimSpace(address))
	} else {
		return nil, errors.BadRequest("INVALID_USER", "缺少 userId 或 address")
	}
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound("USER_NOT_FOUND", "用户不存在")
	}
	invitees, err := uc.userRepo.ListDirectInvitees(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	result := make([]*AdminDownlineNode, 0, len(invitees))
	for _, item := range invitees {
		count, _ := uc.userRepo.CountInvitees(ctx, item.ID)
		result = append(result, &AdminDownlineNode{
			User:            item,
			DirectCount:     count,
			RecommendAmount: item.TeamPerf,
		})
	}
	return result, nil
}

func (uc *AdminUsecase) TriggerSettlement(ctx context.Context, tokenString, settlementDate string) error {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return err
	}
	if settlementDate == "" {
		settlementDate = TodaySettlementDate(token.NowChina())
	}
	return uc.settlement.ForceDailySettlement(ctx, settlementDate)
}

func (uc *AdminUsecase) AdminCreditBalance(ctx context.Context, tokenString, address, amount string) (string, string, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return "", "", err
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return "", "", errors.BadRequest("INVALID_ADDRESS", "请填写用户地址")
	}
	amountDec, err := ParseAmount(strings.TrimSpace(amount))
	if err != nil || !amountDec.GreaterThan(decimal.Zero) {
		return "", "", errors.BadRequest("INVALID_AMOUNT", "充值金额必须大于0")
	}
	user, err := uc.userRepo.FindByAddress(ctx, address)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.NotFound("USER_NOT_FOUND", "用户不存在，请确认地址已注册登录")
	}
	now := time.Now()
	recharge, err := uc.walletRepo.CreateRecharge(ctx, &Recharge{
		UserID:      user.ID,
		Address:     user.Address,
		FromAddress: user.Address,
		Amount:      amountDec.String(),
		Message:     fmt.Sprintf("admin credit %s USDT", amountDec.String()),
		Status:      RechargeStatusPending,
		ExpireAt:    now.Add(24 * time.Hour),
	})
	if err != nil {
		return "", "", err
	}
	balance, err := uc.walletRepo.ConfirmRechargeCredit(ctx, recharge.ID, fmt.Sprintf("admin-%d", recharge.ID))
	if err != nil {
		return "", "", err
	}
	uc.log.Infof("admin credited %s USDT to %s (usdt_recharge), balance=%s", amountDec.String(), user.Address, balance)
	return balance, amountDec.String(), nil
}

// AdminMoveRechargeToReward 管理端：用户充值钱包 → 奖励钱包
func (uc *AdminUsecase) AdminMoveRechargeToReward(ctx context.Context, tokenString string, userID int64, amount string) (string, string, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return "", "", err
	}
	if userID <= 0 {
		return "", "", errors.BadRequest("INVALID_USER", "用户无效")
	}
	amt, err := ParseAmount(strings.TrimSpace(amount))
	if err != nil || !amt.GreaterThan(decimal.Zero) {
		return "", "", errors.BadRequest("INVALID_AMOUNT", "划转金额必须大于0")
	}
	rechargeBal, rewardBal, err := uc.walletRepo.MoveRechargeToReward(ctx, userID, amt.String())
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			return "", "", errors.BadRequest("INSUFFICIENT_BALANCE", "充值钱包余额不足")
		}
		return "", "", err
	}
	return rechargeBal, rewardBal, nil
}

type AdminSettlementBatch struct {
	*SettlementBatch
	ReleaseTotal string
}

func (uc *AdminUsecase) ListSettlementBatches(ctx context.Context, tokenString string, offset, limit int) ([]*AdminSettlementBatch, int64, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, 0, err
	}
	batches, total, err := uc.settlement.ListBatches(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*AdminSettlementBatch, 0, len(batches))
	for _, b := range batches {
		releaseTotal, err := uc.settlement.SumReleaseForBatch(ctx, b)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, &AdminSettlementBatch{SettlementBatch: b, ReleaseTotal: releaseTotal})
	}
	return out, total, nil
}

func (uc *AdminUsecase) ListOrders(ctx context.Context, tokenString string) ([]*AdminOrderDetail, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	return uc.walletRepo.ListAllOrders(ctx)
}

// ListExchangeRecords 管理端：列出所有 AIX→WIN 兑换记录
func (uc *AdminUsecase) ListExchangeRecords(ctx context.Context, tokenString string) ([]*ExchangeRecord, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	return uc.walletRepo.ListAllExchangeRecords(ctx)
}

// ListAllWithdrawals 管理端：列出所有提现记录（仅 WIN）
func (uc *AdminUsecase) ListAllWithdrawals(ctx context.Context, tokenString string) ([]*Withdrawal, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	return uc.walletRepo.ListAllWithdrawals(ctx)
}

func (uc *AdminUsecase) UpdateOrder(ctx context.Context, tokenString string, update *AdminOrderUpdate) (*AdminOrderDetail, error) {
	if _, err := uc.requireAdmin(ctx, tokenString); err != nil {
		return nil, err
	}
	if update == nil || update.OrderID <= 0 {
		return nil, errors.BadRequest("INVALID_ORDER", "订单 ID 无效")
	}
	if update.Status != "" && update.Status != OrderStatusActive && update.Status != OrderStatusExited && update.Status != "completed" {
		return nil, errors.BadRequest("INVALID_STATUS", "订单状态只能是 active 或 exited")
	}
	order, err := uc.walletRepo.AdminUpdateOrder(ctx, update)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("ORDER_NOT_FOUND", "订单不存在")
		}
		return nil, err
	}
	user, err := uc.userRepo.FindByID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}
	addr := ""
	if user != nil {
		addr = user.Address
	}
	return &AdminOrderDetail{Order: order, UserAddress: addr}, nil
}
