package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend/internal/biz"
	"backend/internal/conf"
	"backend/internal/data"
	authmw "backend/internal/middleware"
	jwtpkg "backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/shopspring/decimal"
)

// AdminLegacyService serves /api/admin_dhb/* compatibility routes for the Vue admin UI.
type AdminLegacyService struct {
	admin      *biz.AdminUsecase
	userRepo   biz.UserRepo
	walletRepo biz.WalletRepo
	data       *data.Data
	authCfg    *conf.AuthConfig
	walletCfg  *conf.WalletConfig
}

func NewAdminLegacyService(
	admin *biz.AdminUsecase,
	userRepo biz.UserRepo,
	walletRepo biz.WalletRepo,
	data *data.Data,
	authCfg *conf.AuthConfig,
	walletCfg *conf.WalletConfig,
) *AdminLegacyService {
	return &AdminLegacyService{
		admin: admin, userRepo: userRepo, walletRepo: walletRepo,
		data: data, authCfg: authCfg, walletCfg: walletCfg,
	}
}

var legacyMenuPaths = []string{
	"/home", "/member", "/recharge", "/subscription",
	"/ordersList", "/config", "/settlement", "/lookChildren",
}

type legacyConfigItem struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var legacyConfigDefs = []struct {
	ID   int
	Name string
}{
	{1, "最低认购(U)"},
	{2, "静态利率(%)"},
	{3, "出局倍数"},
	{4, "直推比例"},
	{5, "收款地址"},
	{6, "USDT合约"},
	{7, "AIX价格(USDT/枚)"},
	{8, "WIN价格(USDT/枚)"},
	{9, "兑换手续费率(%)"},
	{31, "USDT充值最小值"},
	{32, "WIN充值最小值"},
	{11, "W1 收益系数"},
	{12, "W2 收益系数"},
	{13, "W3 收益系数"},
	{14, "W4 收益系数"},
	{15, "W5 收益系数"},
	{16, "W6 收益系数"},
	{17, "W7 收益系数"},
	{18, "W8 收益系数"},
	{19, "W9 收益系数"},
	{20, "W10 收益系数"},
	{21, "成为 W1 的小区业绩金额(USDT)"},
	{22, "成为 W2 的小区业绩金额(USDT)"},
	{23, "成为 W3 的小区业绩金额(USDT)"},
	{24, "成为 W4 的小区业绩金额(USDT)"},
	{25, "成为 W5 的小区业绩金额(USDT)"},
	{26, "成为 W6 的小区业绩金额(USDT)"},
	{27, "成为 W7 的小区业绩金额(USDT)"},
	{28, "成为 W8 的小区业绩金额(USDT)"},
	{29, "成为 W9 的小区业绩金额(USDT)"},
	{30, "成为 W10 的小区业绩金额(USDT)"},
}

func (s *AdminLegacyService) HandleLogin(ctx khttp.Context) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	account := firstNonEmpty(
		ctx.Request().Form.Get("username"),
		ctx.Request().Form.Get("account"),
		ctx.Request().Form.Get("email"),
	)
	password := ctx.Request().Form.Get("password")
	if account != s.authCfg.GetAdminAccount() || password != s.authCfg.GetAdminPassword() {
		return errors.Unauthorized("UNAUTHORIZED", "账号或密码错误")
	}
	token, _, err := jwtpkg.Generate(biz.ZeroAddress, s.authCfg.GetJwtSecret(), time.Now())
	if err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"token": token})
}

func (s *AdminLegacyService) HandleMyAuthList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	auth := make([]map[string]string, 0, len(legacyMenuPaths))
	for _, p := range legacyMenuPaths {
		auth = append(auth, map[string]string{"path": p})
	}
	return ctx.Result(200, map[string]interface{}{
		"super": "1",
		"auth":  auth,
	})
}

func (s *AdminLegacyService) HandleAll(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	stats, err := s.buildDashboardStats(ctx)
	if err != nil {
		return err
	}
	return ctx.Result(200, stats)
}

func (s *AdminLegacyService) HandleUserList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))

	if err := s.userRepo.RefreshPerformance(ctx); err != nil {
		return err
	}
	users, err := s.userRepo.ListAllUsers(ctx)
	if err != nil {
		return err
	}
	activeStake, err := s.sumActivePrincipalByUser(ctx)
	if err != nil {
		return err
	}
	directCount := map[int64]int{}
	for _, u := range users {
		if u.InviterID != nil {
			directCount[*u.InviterID]++
		}
	}

	filtered := make([]*biz.User, 0, len(users))
	for _, u := range users {
		if addressFilter != "" && !strings.Contains(strings.ToLower(u.Address), strings.ToLower(addressFilter)) {
			continue
		}
		filtered = append(filtered, u)
	}
	total := len(filtered)
	pageUsers := paginateSlice(filtered, offset, pageSize)

	items := make([]map[string]interface{}, 0, len(pageUsers))
	for _, u := range pageUsers {
		u.SyncCompatFields()
		inviteeCount := directCount[u.ID]
		vip := formatMgmtVIP(u.MgmtLevel)
		items = append(items, map[string]interface{}{
			"userId":              u.ID,
			"id":                  u.ID,
			"address":             u.Address,
			"usdt_recharge":       u.UsdtRecharge,
			"usdt_reward":         u.UsdtReward,
			"aix_balance":         u.AixBalance,        // AIX 代币数
			"win_balance":         u.WinBalance,        // WIN 代币数
			"pending_mgmt_reward": u.OverflowReward, // 兼容旧字段
			"overflow_reward":     u.OverflowReward, // 溢出奖励
			"static_usdt_total":   u.StaticUsdtTotal,   // 静态总收益 USDT
			"mgmt_level":          u.MgmtLevel,
			"large_area_perf":     u.LargeAreaPerf,
			"small_area_perf":     u.SmallAreaPerf,
			"team_perf":           u.TeamPerf,
			"invitee_count":       inviteeCount,
			"createdAt":           formatLegacyTime(u.CreatedAt),
			"myRecommendAddress":  u.InviterAddress,
			"bAmount":             u.UsdtRecharge,
			"amountUsdtCurrent":   zeroIfEmpty(activeStake[u.ID]),
			"vip":                 vip,
			"historyRecommend":    strconv.Itoa(inviteeCount),
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"users": items,
		"count": total,
		"page":  page,
	})
}

func (s *AdminLegacyService) HandleConfig(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	cfg := s.admin.GetPersistedConfigSnapshot()
	items := make([]legacyConfigItem, 0, len(legacyConfigDefs))
	for _, def := range legacyConfigDefs {
		items = append(items, legacyConfigItem{
			ID: def.ID, Name: def.Name, Value: legacyConfigValue(cfg, s.walletCfg, def.ID),
		})
	}
	return ctx.Result(200, map[string]interface{}{"config": items})
}

func (s *AdminLegacyService) HandleConfigUpdate(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	id, _ := strconv.Atoi(ctx.Request().Form.Get("id"))
	value := ctx.Request().Form.Get("value")
	snapshot := s.admin.GetPersistedConfigSnapshot()
	if err := applyLegacyConfigUpdate(snapshot, s.walletCfg, id, value); err != nil {
		return err
	}
	if _, err := s.admin.UpdateSystemConfig(ctx, s.token(ctx), snapshot); err != nil {
		return err
	}
	if id >= 21 && id <= 30 {
		if err := s.userRepo.RefreshPerformance(ctx); err != nil {
			return err
		}
	}
	if id == 7 && strings.TrimSpace(value) != "" {
		date := jwtpkg.NowChina().Format("2006-01-02")
		_ = s.walletRepo.UpsertAixPrice(ctx, date, strings.TrimSpace(value), "admin config")
	}
	if id == 8 && strings.TrimSpace(value) != "" {
		_ = s.walletRepo.UpsertCurrentWinPrice(ctx, strings.TrimSpace(value), "admin")
	}
	return ctx.Result(200, map[string]string{"status": "ok"})
}

func (s *AdminLegacyService) HandleBuyList(ctx khttp.Context) error {
	return s.handleOrderList(ctx)
}

func (s *AdminLegacyService) HandleWithdrawList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))

	list, err := s.admin.ListAllWithdrawals(ctx, s.token(ctx))
	if err != nil {
		return err
	}
	filtered := make([]*biz.Withdrawal, 0, len(list))
	for _, w := range list {
		if addressFilter != "" && !strings.Contains(strings.ToLower(w.Address), strings.ToLower(addressFilter)) {
			continue
		}
		filtered = append(filtered, w)
	}
	total := len(filtered)
	pageItems := paginateSlice(filtered, offset, pageSize)
	items := make([]map[string]interface{}, 0, len(pageItems))
	for _, w := range pageItems {
		items = append(items, map[string]interface{}{
			"id":        w.ID,
			"address":   w.Address,
			"toAddress": w.ToAddress,
			"amount":    w.Amount,
			"fee":       w.Fee,
			"netAmount": w.NetAmount,
			"asset":     w.Asset,
			"status":    w.Status,
			"txHash":    w.TxHash,
			"createdAt": formatLegacyTime(w.CreatedAt),
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"withdraw": items,
		"list":     items,
		"count":    total,
		"page":     page,
	})
}

// HandleExchangeList 管理端：AIX→WIN 兑换记录列表
func (s *AdminLegacyService) HandleExchangeList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))

	list, err := s.admin.ListExchangeRecords(ctx, s.token(ctx))
	if err != nil {
		return err
	}
	filtered := make([]*biz.ExchangeRecord, 0, len(list))
	for _, r := range list {
		if addressFilter != "" && !strings.Contains(strings.ToLower(r.UserAddress), strings.ToLower(addressFilter)) {
			continue
		}
		filtered = append(filtered, r)
	}
	total := len(filtered)
	pageItems := paginateSlice(filtered, offset, pageSize)
	items := make([]map[string]interface{}, 0, len(pageItems))
	for _, r := range pageItems {
		items = append(items, map[string]interface{}{
			"id":            r.ID,
			"address":       r.UserAddress,
			"fromAsset":     r.FromAsset,
			"fromAmount":    r.FromAmount,
			"toAsset":       r.ToAsset,
			"toAmount":      r.ToAmount,
			"feeAmount":     r.FeeAmount,
			"feeRate":       r.FeeRate,
			"exchangePrice": r.ExchangePrice,
			"status":        r.Status,
			"remark":        r.Remark,
			"createdAt":     formatLegacyTime(r.CreatedTime),
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"list":  items,
		"count": total,
		"page":  page,
	})
}

func (s *AdminLegacyService) HandleWithdrawPass(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok", "message": "not supported in AIX"})
}

func (s *AdminLegacyService) HandleRewardList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))
	typeFilter := strings.TrimSpace(firstNonEmpty(q.Get("type"), q.Get("reason")))

	type row struct {
		ID             int64
		Type           string
		Asset          string
		Amount         decimal.Decimal
		Address        string
		FromAddress    string
		SettlementDate *string
		CreatedTime    time.Time
	}
	var rows []row
	db := s.data.DB().WithContext(ctx).
		Table("reward_logs rl").
		Select(`rl.id, rl.type, rl.asset, rl.amount, u.address,
			COALESCE(fu.address,'') as from_address, rl.settlement_date, rl.created_time`).
		Joins("JOIN users u ON u.id = rl.user_id").
		Joins("LEFT JOIN users fu ON fu.id = rl.from_user_id").
		Order("rl.id desc")
	if addressFilter != "" {
		db = db.Where("u.address LIKE ?", "%"+addressFilter+"%")
	}
	if typeFilter != "" && typeFilter != "undefined" && typeFilter != "null" {
		db = db.Where("rl.type = ?", typeFilter)
	}
	if err := db.Scan(&rows).Error; err != nil {
		return err
	}
	total := len(rows)
	pageRows := paginateSlice(rows, offset, pageSize)
	items := make([]map[string]interface{}, 0, len(pageRows))
	for _, r := range pageRows {
		settle := ""
		if r.SettlementDate != nil {
			settle = *r.SettlementDate
		}
		items = append(items, map[string]interface{}{
			"id":             r.ID,
			"type":           r.Type,
			"asset":          r.Asset,
			"amount":         r.Amount.String(),
			"address":        r.Address,
			"addressTwo":     r.FromAddress,
			"reason":         r.Type,
			"settlementDate": settle,
			"createdAt":      formatLegacyTime(r.CreatedTime),
			"date":           formatLegacyTime(r.CreatedTime),
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"rewards": items,
		"list":    items,
		"count":   total,
		"page":    page,
	})
}

func (s *AdminLegacyService) HandleRecordList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))
	typeFilter := strings.ToLower(strings.TrimSpace(q.Get("type"))) // admin | usdt | win

	type row struct {
		ID          int64
		Address     string
		Asset       string
		Amount      decimal.Decimal
		TxHash      string
		Message     string
		CreatedTime time.Time
	}
	var rows []row
	db := s.data.DB().WithContext(ctx).
		Table("recharges r").
		Select(`r.id, COALESCE(NULLIF(r.from_address,''), u.address) as address,
			r.asset, r.amount, r.tx_hash, COALESCE(r.message,'') as message, r.created_time`).
		Joins("JOIN users u ON u.id = r.user_id").
		Where("r.status = ?", biz.RechargeStatusConfirmed).
		Order("r.id desc")
	if addressFilter != "" {
		db = db.Where("(r.from_address LIKE ? OR u.address LIKE ?)", "%"+addressFilter+"%", "%"+addressFilter+"%")
	}
	switch typeFilter {
	case "admin", "后台充值":
		db = db.Where("r.tx_hash LIKE ?", "admin-%")
	case "win", "win充值", "win_recharge":
		db = db.Where("UPPER(r.asset) = ? AND r.tx_hash NOT LIKE ?", biz.TokenWIN, "admin-%")
	case "usdt", "usdt充值", "usdt_recharge":
		db = db.Where("(UPPER(r.asset) = ? OR r.asset = '' OR r.asset IS NULL) AND r.tx_hash NOT LIKE ?", biz.TokenUSDT, "admin-%")
	}
	if err := db.Scan(&rows).Error; err != nil {
		return err
	}
	total := len(rows)
	pageRows := paginateSlice(rows, offset, pageSize)
	items := make([]map[string]interface{}, 0, len(pageRows))
	for _, r := range pageRows {
		remark, typeCode := classifyRechargeType(r.Asset, r.TxHash, r.Message)
		items = append(items, map[string]interface{}{
			"id":        r.ID,
			"address":   r.Address,
			"asset":     strings.ToUpper(strings.TrimSpace(r.Asset)),
			"amount":    r.Amount.String(),
			"txHash":    r.TxHash,
			"remark":    remark,
			"type":      typeCode,
			"createdAt": formatLegacyTime(r.CreatedTime),
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"rewards":   items,
		"list":      items,
		"locations": items,
		"count":     total,
		"page":      page,
	})
}

func classifyRechargeType(asset, txHash, message string) (remark, typeCode string) {
	txHash = strings.TrimSpace(txHash)
	asset = strings.ToUpper(strings.TrimSpace(asset))
	message = strings.ToLower(strings.TrimSpace(message))
	if strings.HasPrefix(txHash, "admin-") {
		return "后台充值", "admin"
	}
	if asset == biz.TokenWIN || strings.Contains(message, "win_deposit") || strings.Contains(message, "win_recharge") {
		return "WIN充值", "win"
	}
	return "USDT充值", "usdt"
}

func (s *AdminLegacyService) HandleAdminRecharge(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	address := strings.TrimSpace(ctx.Request().Form.Get("address"))
	amount := strings.TrimSpace(ctx.Request().Form.Get("amount"))
	balance, credited, err := s.admin.AdminCreditBalance(ctx, s.token(ctx), address, amount)
	if err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{
		"status":  "ok",
		"balance": balance,
		"amount":  credited,
		"message": "充值成功",
	})
}

func (s *AdminLegacyService) HandleAdminRechargeWin(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	address := strings.TrimSpace(ctx.Request().Form.Get("address"))
	amount := strings.TrimSpace(ctx.Request().Form.Get("amount"))
	balance, credited, err := s.admin.AdminCreditWinBalance(ctx, s.token(ctx), address, amount)
	if err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{
		"status":      "ok",
		"asset":       "WIN",
		"win_balance": balance,
		"amount":      credited,
		"message":     "WIN 充值成功",
	})
}

func (s *AdminLegacyService) HandleRechargeToReward(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	userID, _ := strconv.ParseInt(ctx.Request().Form.Get("user_id"), 10, 64)
	amount := strings.TrimSpace(ctx.Request().Form.Get("amount"))
	rechargeBal, rewardBal, err := s.admin.AdminMoveRechargeToReward(ctx, s.token(ctx), userID, amount)
	if err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{
		"status":        "ok",
		"usdt_recharge": rechargeBal,
		"usdt_reward":   rewardBal,
		"message":       "已转入奖励钱包",
	})
}

func (s *AdminLegacyService) HandleGoodList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{
		"goods": []interface{}{},
		"count": 0,
		"page":  1,
	})
}

func (s *AdminLegacyService) HandleStubGoods(ctx khttp.Context) error {
	return s.HandleGoodList(ctx)
}

func (s *AdminLegacyService) HandleLockUserReward(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	form := ctx.Request().Form
	if form.Get("settlement_date") != "" || form.Get("trigger_settlement") == "1" {
		return s.triggerSettlement(ctx, form.Get("settlement_date"))
	}
	return ctx.Result(200, map[string]string{"status": "ok"})
}

func (s *AdminLegacyService) HandleLocationInsert(ctx khttp.Context) error {
	return s.HandleLockUserReward(ctx)
}

func (s *AdminLegacyService) HandleSettlementList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	batches, total, err := s.admin.ListSettlementBatches(ctx, s.token(ctx), offset, pageSize)
	if err != nil {
		return err
	}
	list := make([]map[string]interface{}, 0, len(batches))
	for _, b := range batches {
		status := b.Status
		if status == biz.SettlementStatusSuccess {
			status = "completed"
		}
		item := map[string]interface{}{
			"id":             b.ID,
			"settlementDate": b.SettlementDate,
			"status":         status,
			"releaseTotal":   b.ReleaseTotal,
			"aixPrice":       b.AixPrice,
			"staticAmount":   b.StaticAmount,
			"mgmtAmount":     b.MgmtAmount,
			"startedAt":      "",
			"finishedAt":     "",
		}
		if !b.StartedAt.IsZero() {
			item["startedAt"] = b.StartedAt.In(jwtpkg.ChinaLocation()).Format("2006-01-02 15:04:05")
		}
		if b.FinishedAt != nil {
			item["finishedAt"] = b.FinishedAt.In(jwtpkg.ChinaLocation()).Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}
	return ctx.Result(200, map[string]interface{}{
		"list":              list,
		"total":             total,
		"count":             total,
		"page":              page,
		"defaultSettleDate": biz.TodaySettlementDate(jwtpkg.NowChina()),
	})
}

func (s *AdminLegacyService) HandleSettlementTrigger(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	date := ""
	_ = ctx.Request().ParseForm()
	date = firstNonEmpty(
		ctx.Request().Form.Get("settlement_date"),
		ctx.Request().Form.Get("date"),
		ctx.Request().URL.Query().Get("date"),
	)
	if date == "" {
		body, _ := io.ReadAll(ctx.Request().Body)
		var m map[string]string
		if json.Unmarshal(body, &m) == nil {
			date = firstNonEmpty(m["settlement_date"], m["date"])
		}
	}
	return s.triggerSettlement(ctx, date)
}

func (s *AdminLegacyService) HandleVipUpdate(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := ctx.Request().ParseForm(); err != nil {
		return errors.BadRequest("INVALID_FORM", "请求格式错误")
	}
	userID, _ := strconv.ParseInt(firstNonEmpty(
		ctx.Request().Form.Get("user_id"),
		ctx.Request().Form.Get("userId"),
	), 10, 64)
	vipRaw := firstNonEmpty(
		ctx.Request().Form.Get("vip"),
		ctx.Request().Form.Get("level"),
	)
	level := parseMgmtLevel(vipRaw)
	if level < 0 || level > 10 {
		return errors.BadRequest("INVALID_VIP", "级别须为 W0–W10")
	}
	if _, err := s.admin.UpdateUser(ctx, s.token(ctx), &biz.AdminUserUpdate{
		UserID: userID, CommunityLevel: formatMgmtVIP(level), SetCommunityLevel: true,
	}); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok"})
}

func (s *AdminLegacyService) HandleUpdateGoods(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok", "message": "not supported in AIX"})
}

func (s *AdminLegacyService) HandleUploadGoods(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok", "message": "not supported in AIX"})
}

func (s *AdminLegacyService) HandleStubOK(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok"})
}

func (s *AdminLegacyService) HandleStubRewards(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{"rewards": []interface{}{}, "count": 0})
}

func (s *AdminLegacyService) HandleStubLocations(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return ctx.Result(200, map[string]interface{}{"locations": []interface{}{}, "count": 0})
}

func (s *AdminLegacyService) HandleUserRecommend(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	userID, _ := strconv.ParseInt(firstNonEmpty(q.Get("userId"), q.Get("user_id")), 10, 64)
	address := strings.TrimSpace(firstNonEmpty(q.Get("address"), q.Get("Address")))
	nodes, err := s.admin.ListUserDownline(ctx, s.token(ctx), userID, address)
	if err != nil {
		return err
	}
	users := make([]map[string]interface{}, 0, len(nodes))
	for _, n := range nodes {
		if n == nil || n.User == nil {
			continue
		}
		n.User.SyncCompatFields()
		users = append(users, map[string]interface{}{
			"userId":             n.User.ID,
			"address":            n.User.Address,
			"createdAt":          formatLegacyTime(n.User.CreatedAt),
			"amount":             n.User.UsdtRecharge,
			"recommendAllAmount": n.RecommendAmount,
			"recommendNum":       n.DirectCount,
			"vip":                formatMgmtVIP(n.User.MgmtLevel),
			"mgmt_level":         n.User.MgmtLevel,
			"team_perf":          n.User.TeamPerf,
			"teamStake":          n.User.TeamPerf,
		})
	}
	return ctx.Result(200, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func (s *AdminLegacyService) handleOrderList(ctx khttp.Context) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	q := ctx.Request().URL.Query()
	page, pageSize, offset := parsePage(q)
	addressFilter := strings.TrimSpace(q.Get("address"))
	statusFilter := strings.ToLower(strings.TrimSpace(q.Get("status")))
	if statusFilter == "undefined" || statusFilter == "null" {
		statusFilter = ""
	}
	if statusFilter == "completed" {
		statusFilter = biz.OrderStatusExited
	}
	if strings.EqualFold(addressFilter, "undefined") || strings.EqualFold(addressFilter, "null") {
		addressFilter = ""
	}

	orders, err := s.walletRepo.ListAllOrders(ctx)
	if err != nil {
		return err
	}
	filtered := make([]*biz.AdminOrderDetail, 0, len(orders))
	for _, o := range orders {
		if addressFilter != "" && !strings.Contains(strings.ToLower(o.UserAddress), strings.ToLower(addressFilter)) {
			continue
		}
		if statusFilter != "" && strings.ToLower(strings.TrimSpace(o.Order.Status)) != statusFilter {
			continue
		}
		filtered = append(filtered, o)
	}
	total := len(filtered)
	pageItems := paginateSlice(filtered, offset, pageSize)
	rewards := make([]map[string]interface{}, 0, len(pageItems))
	for _, o := range pageItems {
		rewards = append(rewards, mapLegacyBuyOrder(o))
	}
	return ctx.Result(200, map[string]interface{}{
		"rewards":  rewards,
		"count":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (s *AdminLegacyService) triggerSettlement(ctx khttp.Context, settlementDate string) error {
	if err := s.admin.TriggerSettlement(ctx, s.token(ctx), settlementDate); err != nil {
		return err
	}
	return ctx.Result(200, map[string]string{"status": "ok", "message": "结算任务已触发"})
}

func (s *AdminLegacyService) requireAdmin(ctx khttp.Context) (*biz.User, error) {
	token := s.token(ctx)
	if token == "" {
		return nil, errors.Unauthorized("UNAUTHORIZED", "请先登录")
	}
	return s.admin.RequireAdminUser(ctx, token)
}

func (s *AdminLegacyService) token(ctx khttp.Context) string {
	if t := authmw.ResolveToken(ctx, ""); t != "" {
		return t
	}
	if t := authmw.ParseBearer(ctx.Request().Header.Get("Authorization")); t != "" {
		return t
	}
	if t := strings.TrimSpace(ctx.Request().Header.Get("Access-Token")); t != "" {
		return t
	}
	return strings.TrimSpace(ctx.Request().Header.Get("token"))
}

func (s *AdminLegacyService) buildDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now().In(jwtpkg.ChinaLocation())
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jwtpkg.ChinaLocation())
	todayDate := todayStart.Format("2006-01-02")

	var totalUserR int64
	var todayUserR int64
	if err := s.data.DB().WithContext(ctx).Model(&data.UserPO{}).Count(&totalUserR).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Model(&data.UserPO{}).
		Where("created_time >= ?", todayStart).Count(&todayUserR).Error; err != nil {
		return nil, err
	}

	var totalUser int64
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Select("COUNT(DISTINCT user_id)").Scan(&totalUser).Error; err != nil {
		return nil, err
	}
	var todayUser int64
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Where("created_time >= ?", todayStart).
		Select("COUNT(DISTINCT user_id)").Scan(&todayUser).Error; err != nil {
		return nil, err
	}

	var buyTotal decimal.Decimal
	var todayBuy decimal.Decimal
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Select("COALESCE(SUM(principal),0)").Scan(&buyTotal).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Where("created_time >= ?", todayStart).
		Select("COALESCE(SUM(principal),0)").Scan(&todayBuy).Error; err != nil {
		return nil, err
	}

	var balanceUsdt decimal.Decimal
	var usdtRechargeTotal decimal.Decimal
	var aixTotal decimal.Decimal
	if err := s.data.DB().WithContext(ctx).Model(&data.UserPO{}).
		Select("COALESCE(SUM(usdt_reward),0)").Scan(&balanceUsdt).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Model(&data.UserPO{}).
		Select("COALESCE(SUM(usdt_recharge),0)").Scan(&usdtRechargeTotal).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Model(&data.UserPO{}).
		Select("COALESCE(SUM(aix_balance),0)").Scan(&aixTotal).Error; err != nil {
		return nil, err
	}

	var orderActive int64
	var orderExited int64
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Where("status = ?", biz.OrderStatusActive).Count(&orderActive).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Table("orders").
		Where("status = ?", biz.OrderStatusExited).Count(&orderExited).Error; err != nil {
		return nil, err
	}

	var todayOne decimal.Decimal
	var todayTwo decimal.Decimal
	if err := s.data.DB().WithContext(ctx).Table("reward_logs").
		Where("type = ? AND settlement_date = ?", biz.RewardTypeStaticAix, todayDate).
		Select("COALESCE(SUM(amount),0)").Scan(&todayOne).Error; err != nil {
		return nil, err
	}
	if err := s.data.DB().WithContext(ctx).Table("reward_logs").
		Where("type IN ? AND DATE(created_time) = ?",
			[]string{biz.RewardTypeDynamicUsdt, biz.RewardTypeMgmt}, todayDate).
		Select("COALESCE(SUM(amount),0)").Scan(&todayTwo).Error; err != nil {
		return nil, err
	}
	todayThree := todayOne.Add(todayTwo)

	var totalReward decimal.Decimal
	if err := s.data.DB().WithContext(ctx).Table("reward_logs").
		Where("type IN ?", []string{biz.RewardTypeStaticAix, biz.RewardTypeDynamicUsdt, biz.RewardTypeMgmt}).
		Select("COALESCE(SUM(amount),0)").Scan(&totalReward).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalUserR":        totalUserR,
		"totalUser":         totalUser,
		"todayUserR":        todayUserR,
		"todayUser":         todayUser,
		"buyTotal":          buyTotal.String(),
		"todayBuy":          todayBuy.String(),
		"balanceUsdt":       balanceUsdt.String(),
		"usdtRechargeTotal": usdtRechargeTotal.String(),
		"aixTotal":          aixTotal.String(),
		"orderActive":       orderActive,
		"orderExited":       orderExited,
		"todayOne":          todayOne.String(),
		"todayTwo":          todayTwo.String(),
		"todayThree":        todayThree.String(),
		"totalReward":       totalReward.String(),
		"todayWithdraw":     "0",
		"totalWithdraw":     "0",
		"totalIspay":        "0",
	}, nil
}

func (s *AdminLegacyService) sumActivePrincipalByUser(ctx context.Context) (map[int64]string, error) {
	type row struct {
		UserID int64
		Total  decimal.Decimal
	}
	var rows []row
	err := s.data.DB().WithContext(ctx).Table("orders").
		Select("user_id, COALESCE(SUM(principal),0) as total").
		Where("status = ?", biz.OrderStatusActive).
		Group("user_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]string, len(rows))
	for _, r := range rows {
		result[r.UserID] = r.Total.String()
	}
	return result, nil
}

func mapLegacyBuyOrder(o *biz.AdminOrderDetail) map[string]interface{} {
	principal, err := decimal.NewFromString(o.Order.Principal)
	if err != nil {
		principal = decimal.Zero
	}
	exitCap, err := decimal.NewFromString(o.Order.ExitCap)
	if err != nil {
		exitCap = biz.CalcExitCap(principal)
	}
	earned, err := decimal.NewFromString(o.Order.EarnedTotal)
	if err != nil {
		earned = decimal.Zero
	}
	remain := exitCap.Sub(earned)
	if remain.IsNegative() {
		remain = decimal.Zero
	}
	exitMul := decimal.NewFromFloat(biz.ExitMultiplier)
	if !exitMul.IsPositive() {
		exitMul = decimal.NewFromInt(4)
	}
	status := o.Order.Status
	if status == "completed" {
		status = biz.OrderStatusExited
	}
	return map[string]interface{}{
		"id":          o.Order.ID,
		"address":     o.UserAddress,
		"amount":      principal.String(),
		"exitAmount":  exitMul.String(),
		"money":       exitCap.String(),
		"amountGet":   earned.String(),
		"amountLast":  remain.String(),
		"fund_source": o.Order.FundSource,
		"status":      status,
		"createdAt":   formatLegacyTime(o.Order.CreatedAt),
		"one":         o.Order.FundSource,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func zeroIfEmpty(v string) string {
	if strings.TrimSpace(v) == "" {
		return "0"
	}
	return v
}

func formatMgmtVIP(level int32) string {
	if level < 0 {
		level = 0
	}
	if level > 10 {
		level = 10
	}
	return "W" + strconv.Itoa(int(level))
}

func parseMgmtLevel(raw string) int32 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	raw = strings.TrimPrefix(strings.ToUpper(raw), "W")
	raw = strings.TrimPrefix(raw, "V")
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	if n > 10 {
		return 10
	}
	return int32(n)
}

func formatLegacyTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(jwtpkg.ChinaLocation()).Format("2006-01-02 15:04:05")
}

func parsePage(q url.Values) (page, pageSize, offset int) {
	page = 1
	pageSize = 20
	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := q.Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 1000 {
			pageSize = v
		}
	} else if ps := q.Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 1000 {
			pageSize = v
		}
	}
	offset = (page - 1) * pageSize
	return page, pageSize, offset
}

func paginateSlice[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func legacyConfigValue(cfg *conf.SystemConfigSnapshot, walletCfg *conf.WalletConfig, id int) string {
	if cfg != nil {
		conf.NormalizeBusinessDefaults(cfg)
	}
	switch id {
	case 1:
		if cfg != nil && cfg.MinSubscribe != "" {
			return cfg.MinSubscribe
		}
		if walletCfg != nil {
			return walletCfg.MinSubscribe
		}
		return conf.DefaultMinSubscribe
	case 2:
		if cfg != nil {
			return strconv.FormatFloat(cfg.StaticRate, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultStaticRate, 'f', -1, 64)
	case 3:
		if cfg != nil {
			return strconv.FormatFloat(cfg.ExitMultiplier, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultExitMultiplier, 'f', -1, 64)
	case 4:
		if cfg != nil {
			return strconv.FormatFloat(cfg.DirectRate, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultDirectRate, 'f', -1, 64)
	case 5:
		if cfg != nil {
			if len(cfg.DepositAddresses) > 0 {
				return strings.Join(cfg.DepositAddresses, ",")
			}
			if cfg.DepositAddress != "" {
				return cfg.DepositAddress
			}
		}
		if walletCfg != nil {
			addrs := walletCfg.GetDepositAddresses()
			if len(addrs) > 0 {
				return strings.Join(addrs, ",")
			}
		}
		return ""
	case 6:
		if cfg != nil && cfg.UsdtContract != "" {
			return cfg.UsdtContract
		}
		if walletCfg != nil {
			return walletCfg.UsdtContract
		}
		return ""
	case 7:
		if cfg != nil {
			return strconv.FormatFloat(cfg.AixPriceInitial, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultAixPrice, 'f', -1, 64)
	case 8:
		if cfg != nil {
			return strconv.FormatFloat(cfg.WinPrice, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultWinPrice, 'f', -1, 64)
	case 9:
		if cfg != nil {
			return strconv.FormatFloat(cfg.ExchangeFeeRate*100, 'f', -1, 64)
		}
		return strconv.FormatFloat(conf.DefaultExchangeFeeRate*100, 'f', -1, 64)
	case 31:
		if cfg != nil && strings.TrimSpace(cfg.MinUsdtRecharge) != "" {
			return cfg.MinUsdtRecharge
		}
		return conf.DefaultMinUsdtRecharge
	case 32:
		if cfg != nil && strings.TrimSpace(cfg.MinWinRecharge) != "" {
			return cfg.MinWinRecharge
		}
		return conf.DefaultMinWinRecharge
	case 11, 12, 13, 14, 15, 16, 17, 18, 19, 20:
		idx := id - 11
		rates := conf.DefaultMgmtRates()
		if cfg != nil && len(cfg.MgmtRates) == 10 {
			rates = cfg.MgmtRates
		}
		return strconv.FormatFloat(rates[idx], 'f', -1, 64)
	case 21, 22, 23, 24, 25, 26, 27, 28, 29, 30:
		idx := id - 21
		thresholds := conf.DefaultMgmtThresholds()
		if cfg != nil && len(cfg.MgmtThresholds) == 10 {
			thresholds = cfg.MgmtThresholds
		}
		return strconv.FormatFloat(thresholds[idx], 'f', -1, 64)
	default:
		return ""
	}
}

func applyLegacyConfigUpdate(snapshot *conf.SystemConfigSnapshot, walletCfg *conf.WalletConfig, id int, value string) error {
	if snapshot == nil {
		return errors.BadRequest("INVALID_CONFIG", "配置不可用")
	}
	conf.NormalizeBusinessDefaults(snapshot)
	value = strings.TrimSpace(value)
	switch id {
	case 1:
		snapshot.MinSubscribe = value
	case 2:
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil || rate <= 0 {
			return errors.BadRequest("INVALID_VALUE", "静态利率格式错误")
		}
		snapshot.StaticRate = rate
	case 3:
		mul, err := strconv.ParseFloat(value, 64)
		if err != nil || mul <= 0 {
			return errors.BadRequest("INVALID_VALUE", "出局倍数格式错误")
		}
		snapshot.ExitMultiplier = mul
	case 4:
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil || rate < 0 || rate > 1 {
			return errors.BadRequest("INVALID_VALUE", "直推比例须为 0~1（如 0.5 表示 50%）")
		}
		snapshot.DirectRate = rate
	case 5:
		tmp := &conf.WalletConfig{DepositAddress: value}
		addrs := tmp.GetDepositAddresses()
		snapshot.DepositAddresses = addrs
		if len(addrs) > 0 {
			snapshot.DepositAddress = addrs[0]
		} else {
			snapshot.DepositAddress = value
		}
		_ = walletCfg
	case 6:
		snapshot.UsdtContract = value
	case 7:
		price, err := strconv.ParseFloat(value, 64)
		if err != nil || price <= 0 {
			return errors.BadRequest("INVALID_VALUE", "AIX价格格式错误")
		}
		snapshot.AixPriceInitial = price
	case 8:
		price, err := strconv.ParseFloat(value, 64)
		if err != nil || price <= 0 {
			return errors.BadRequest("INVALID_VALUE", "WIN价格格式错误")
		}
		snapshot.WinPrice = price
	case 9:
		feePercent, err := strconv.ParseFloat(value, 64)
		if err != nil || feePercent < 0 || feePercent >= 100 {
			return errors.BadRequest("INVALID_VALUE", "手续费率须为 0~100 的百分数（如 5 表示 5%）")
		}
		snapshot.ExchangeFeeRate = feePercent / 100.0
	case 31:
		minAmt, err := strconv.ParseFloat(value, 64)
		if err != nil || minAmt < conf.FloorMinRechargeAmount {
			return errors.BadRequest("INVALID_VALUE", fmt.Sprintf("USDT充值最小值必须 ≥ %g", conf.FloorMinRechargeAmount))
		}
		snapshot.MinUsdtRecharge = strconv.FormatFloat(minAmt, 'f', -1, 64)
	case 32:
		minAmt, err := strconv.ParseFloat(value, 64)
		if err != nil || minAmt < conf.FloorMinRechargeAmount {
			return errors.BadRequest("INVALID_VALUE", fmt.Sprintf("WIN充值最小值必须 ≥ %g", conf.FloorMinRechargeAmount))
		}
		snapshot.MinWinRecharge = strconv.FormatFloat(minAmt, 'f', -1, 64)
	case 11, 12, 13, 14, 15, 16, 17, 18, 19, 20:
		rate, err := strconv.ParseFloat(value, 64)
		if err != nil || rate < 0 {
			return errors.BadRequest("INVALID_VALUE", "收益系数格式错误（如 0.2 表示 20%）")
		}
		if len(snapshot.MgmtRates) != 10 {
			snapshot.MgmtRates = conf.DefaultMgmtRates()
		}
		snapshot.MgmtRates[id-11] = rate
	case 21, 22, 23, 24, 25, 26, 27, 28, 29, 30:
		threshold, err := strconv.ParseFloat(value, 64)
		if err != nil || threshold <= 0 {
			return errors.BadRequest("INVALID_VALUE", "晋级业绩金额必须为大于 0 的数字")
		}
		if len(snapshot.MgmtThresholds) != 10 {
			snapshot.MgmtThresholds = conf.DefaultMgmtThresholds()
		}
		idx := id - 21
		if idx > 0 && threshold <= snapshot.MgmtThresholds[idx-1] {
			return errors.BadRequest("INVALID_VALUE", "晋级业绩金额必须高于前一级")
		}
		if idx < len(snapshot.MgmtThresholds)-1 && threshold >= snapshot.MgmtThresholds[idx+1] {
			return errors.BadRequest("INVALID_VALUE", "晋级业绩金额必须低于后一级")
		}
		snapshot.MgmtThresholds[idx] = threshold
	default:
		return errors.BadRequest("INVALID_ID", "未知配置项")
	}
	return nil
}
