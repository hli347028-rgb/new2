package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/biz"
	authmw "backend/internal/middleware"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterWalletExtraRoutes mounts AIX-specific routes without regenerating protos.
func RegisterWalletExtraRoutes(srv *khttp.Server, wallet *WalletService) {
	r := srv.Route("/")
	r.POST("/v1/wallet/subscribe-aix", wallet.HandleSubscribeAIX)
	r.POST("/v1/wallet/transfer", wallet.HandleTransfer)
	r.POST("/v1/wallet/recharge-to-reward", wallet.HandleRechargeToReward)
	r.GET("/v1/wallet/transfer-records/self", wallet.HandleSelfTransferRecords)
	r.GET("/v1/wallet/transfer-records/lineal", wallet.HandleLinealTransferRecords)
	r.POST("/v1/wallet/withdraw-aix", wallet.HandleWithdrawAIX)
	r.POST("/v1/wallet/exchange-aix-to-win", wallet.HandleExchangeAixToWin)
	r.POST("/v1/wallet/withdraw-win", wallet.HandleWithdrawWIN)
	r.GET("/v1/wallet/exchange-records", wallet.HandleExchangeRecords)
	r.GET("/v1/wallet/aix-price", wallet.HandleAixPrice)
	r.GET("/v1/wallet/rewards", wallet.HandleRewards)
	r.GET("/v1/wallet/management-rewards", wallet.HandleManagementRewards)
	r.GET("/v1/wallet/aix-profile", wallet.HandleAixProfile)
	r.POST("/v1/wallet/recharge-win", wallet.HandleCreateWinRecharge)
	r.POST("/v1/wallet/recharge-win/confirm", wallet.HandleConfirmWinRecharge)
	r.GET("/v1/wallet/recharges-win", wallet.HandleListWinRecharges)
}

func transferRecordPagination(ctx khttp.Context) (page, pageSize int, err error) {
	page = 1
	pageSize = 10
	query := ctx.Request().URL.Query()
	if raw := strings.TrimSpace(query.Get("page")); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page <= 0 {
			return 0, 0, fmt.Errorf("page 必须为大于0的整数")
		}
	}
	if raw := strings.TrimSpace(query.Get("page_size")); raw != "" {
		pageSize, err = strconv.Atoi(raw)
		if err != nil || pageSize <= 0 || pageSize > 100 {
			return 0, 0, fmt.Errorf("page_size 必须为1到100的整数")
		}
	}
	return page, pageSize, nil
}

func pageBounds(total, page, pageSize int) (int, int) {
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

// tokenFromRequest 自定义 Route 可能不走 Middleware，需直接从 HTTP 头取 Bearer。
func tokenFromRequest(ctx khttp.Context, fallback string) string {
	if t := resolveToken(ctx, fallback); t != "" {
		return t
	}
	if t := authmw.TokenFromContext(ctx); t != "" {
		return t
	}
	req := ctx.Request()
	if req == nil {
		return strings.TrimSpace(fallback)
	}
	if t := authmw.ParseBearer(req.Header.Get("Authorization")); t != "" {
		return t
	}
	if t := strings.TrimSpace(req.Header.Get("Access-Token")); t != "" {
		return t
	}
	if t := strings.TrimSpace(req.Header.Get("token")); t != "" {
		return t
	}
	if t := strings.TrimSpace(req.URL.Query().Get("token")); t != "" {
		return t
	}
	return strings.TrimSpace(fallback)
}

type subscribeAIXReq struct {
	Token   string `json:"token"`
	Amount  string `json:"amount"`
	PayFrom string `json:"pay_from"`
}

func (s *WalletService) HandleSubscribeAIX(ctx khttp.Context) error {
	var req subscribeAIXReq
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	order, bal, err := s.uc.SubscribeAIX(ctx, token, req.Amount, req.PayFrom)
	if err != nil {
		return err
	}
	resp := map[string]any{
		"order_id":     order.ID,
		"principal":    order.Principal,
		"exit_cap":     order.ExitCap,
		"fund_source":  order.FundSource,
		"direct_base":  order.DirectBase,
		"status":       order.Status,
		"balance":      bal,
		"total_amount": order.Principal,
	}
	if order.FundSource == biz.PayFromWin {
		resp["from_win"] = order.FromWin
		resp["win_price"] = order.WinPrice
		resp["win_balance"] = bal
		if order.WinPrice == "" {
			resp["win_price"] = fmt.Sprintf("%v", biz.GetWinPrice())
		}
	}
	return ctx.JSON(http.StatusOK, resp)
}

type transferReq struct {
	Token     string `json:"token"`
	ToAddress string `json:"to_address"`
	Asset     string `json:"asset"`
	Amount    string `json:"amount"`
	PayFrom   string `json:"pay_from"`
}

func (s *WalletService) HandleTransfer(ctx khttp.Context) error {
	var req transferReq
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	t, err := s.uc.Transfer(ctx, token, req.ToAddress, req.Asset, req.Amount, req.PayFrom)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"id": t.ID, "from_user_id": t.FromUserID, "to_user_id": t.ToUserID,
		"asset": t.Asset, "amount": t.Amount, "pay_from": t.PayFrom,
		"to_credit_reward": t.ToCreditReward, "to_credit_aix": t.ToCreditAix,
	})
}

func (s *WalletService) HandleRechargeToReward(ctx khttp.Context) error {
	var req struct {
		Token  string `json:"token"`
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	rechargeBal, rewardBal, err := s.uc.MoveRechargeToReward(ctx, token, req.Amount)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"usdt_recharge": rechargeBal,
		"usdt_reward":   rewardBal,
	})
}

func (s *WalletService) HandleSelfTransferRecords(ctx khttp.Context) error {
	page, pageSize, err := transferRecordPagination(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"code": 400, "reason": "INVALID_PAGE", "message": err.Error(),
		})
	}
	records, err := s.uc.ListSelfTransferRecords(ctx, tokenFromRequest(ctx, ""))
	if err != nil {
		return err
	}
	start, end := pageBounds(len(records), page, pageSize)
	list := make([]map[string]any, 0, end-start)
	for _, record := range records[start:end] {
		list = append(list, map[string]any{
			"id": record.ID, "asset": record.Asset, "amount": record.Amount,
			"from_wallet": record.FromWallet, "to_wallet": record.ToWallet,
			"created_at": record.CreatedTime.Unix(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"page": page, "page_size": pageSize, "total": len(records), "list": list,
	})
}

func (s *WalletService) HandleLinealTransferRecords(ctx khttp.Context) error {
	page, pageSize, err := transferRecordPagination(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"code": 400, "reason": "INVALID_PAGE", "message": err.Error(),
		})
	}
	direction := strings.ToLower(strings.TrimSpace(ctx.Request().URL.Query().Get("direction")))
	records, err := s.uc.ListLinealTransferRecords(ctx, tokenFromRequest(ctx, ""), direction)
	if err != nil {
		return err
	}
	start, end := pageBounds(len(records), page, pageSize)
	list := make([]map[string]any, 0, end-start)
	for _, record := range records[start:end] {
		list = append(list, map[string]any{
			"id": record.ID, "direction": record.Direction, "relationship": record.Relationship,
			"counterparty_user_id": record.CounterpartyUserID,
			"counterparty_address": record.CounterpartyAddress,
			"asset":                record.Asset, "amount": record.Amount,
			"from_wallet": record.FromWallet, "to_wallet": record.ToWallet,
			"created_at": record.CreatedTime.Unix(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"page": page, "page_size": pageSize, "total": len(records), "list": list,
	})
}

func (s *WalletService) HandleWithdrawAIX(ctx khttp.Context) error {
	var req struct {
		Token     string `json:"token"`
		Amount    string `json:"amount"`
		ToAddress string `json:"to_address"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	w, left, err := s.uc.CreateAixWithdraw(ctx, token, req.Amount, req.ToAddress)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"withdraw_id":  w.ID,
		"asset":        "AIX",
		"amount":       w.Amount,
		"to_address":   w.ToAddress,
		"status":       w.Status,
		"tx_hash":      w.TxHash, // 合约未就绪前为空
		"aix_balance":  left,
		"aix_contract": "", // TODO: 待配置 AIX 代币合约
	})
}

func (s *WalletService) HandleAixPrice(ctx khttp.Context) error {
	date := ctx.Request().URL.Query().Get("date")
	price, err := s.uc.GetAixPrice(ctx, date)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"price": price, "date": date,
		"aix_contract": "", // TODO: 待配置 AIX 代币合约
	})
}

func (s *WalletService) HandleRewards(ctx khttp.Context) error {
	token := tokenFromRequest(ctx, "")
	logs, err := s.uc.ListRewardLogs(ctx, token)
	if err != nil {
		return err
	}
	items := make([]map[string]any, 0, len(logs))
	for _, l := range logs {
		items = append(items, map[string]any{
			"id": l.ID, "type": l.Type, "asset": l.Asset, "amount": l.Amount,
			"base_amount": l.BaseAmount, "rate": l.Rate, "exit_applied": l.ExitApplied,
			"settlement_date": l.SettlementDate, "created_time": l.CreatedTime.Unix(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]any{"rewards": items})
}

func (s *WalletService) HandleManagementRewards(ctx khttp.Context) error {
	token := tokenFromRequest(ctx, "")
	rewards, err := s.uc.ListMgmtRewards(ctx, token)
	if err != nil {
		return err
	}
	items := make([]map[string]any, 0, len(rewards))
	for _, reward := range rewards {
		sourceAddress := ""
		if address, err := s.uc.FindUserAddress(ctx, reward.FromUserID); err == nil {
			sourceAddress = address
		}
		items = append(items, map[string]any{
			"id": reward.ID, "source_user_id": reward.FromUserID,
			"source_address": sourceAddress, "source_order_id": reward.SourceOrderID,
			"base_amount": reward.BaseAmount, "rate": reward.Rate,
			"total_amount": reward.TotalAmount, "released_amount": reward.ReleasedAmount,
			"pending_amount": reward.PendingAmount, "created_time": reward.CreatedTime.Unix(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]any{"rewards": items})
}

// HandleAixProfile 用户端资产总览（含静态总收益）
func (s *WalletService) HandleAixProfile(ctx khttp.Context) error {
	token := tokenFromRequest(ctx, "")
	user, summary, recharge, reward, pending, aix, unexited, nextReleaseAt, serverTime, err := s.uc.GetBalance(ctx, token)
	if err != nil {
		return err
	}
	totalNodes := int32(0)
	if summary != nil {
		totalNodes = summary.TotalNodes
	}
	staticTotal := "0"
	if user != nil && user.StaticUsdtTotal != "" {
		staticTotal = user.StaticUsdtTotal
	}
	mgmtLevel := int32(0)
	largeArea := "0"
	smallArea := "0"
	teamPerf := "0"
	if user != nil {
		mgmtLevel = user.MgmtLevel
		if user.LargeAreaPerf != "" {
			largeArea = user.LargeAreaPerf
		}
		if user.SmallAreaPerf != "" {
			smallArea = user.SmallAreaPerf
		}
		if user.TeamPerf != "" {
			teamPerf = user.TeamPerf
		}
	}
	mgmtSummary, err := s.uc.GetMgmtRewardSummary(ctx, token)
	if err != nil {
		return err
	}

	// AIX 现价以管理配置 AixPriceInitial 为准（当日 aix_prices 供日结历史用）
	aixPrice := biz.AixPriceInitial
	if aixPrice <= 0 {
		if priceStr, priceErr := s.uc.GetAixPrice(ctx, ""); priceErr == nil {
			if parsed, parseErr := strconv.ParseFloat(strings.TrimSpace(priceStr), 64); parseErr == nil && parsed > 0 {
				aixPrice = parsed
			}
		}
	}
	winPrice := biz.GetWinPrice()
	aixToWinRate := 0.0
	if winPrice > 0 {
		// 1 AIX 可兑多少 WIN（毛量，未扣手续费）= aix_price / win_price
		aixToWinRate = aixPrice / winPrice
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"address":              user.Address,
		"usdt_recharge":        recharge,
		"usdt_reward":          reward,
		"aix_balance":          aix,
		"win_balance":          user.WinBalance,
		"pending_mgmt_reward":  user.OverflowReward, // 兼容旧字段，值为溢出奖励
		"overflow_reward":      user.OverflowReward,
		"static_usdt_total":    staticTotal,
		"pending_amount":       pending,
		"unexited_amount":      unexited,
		"total_nodes":          totalNodes,
		"mgmt_level":           mgmtLevel,
		"mgmt_reward_released": mgmtSummary.Released,
		"mgmt_reward_pending":  mgmtSummary.Pending,
		"mgmt_reward_total":    mgmtSummary.Total,
		"large_area_perf":      largeArea,
		"small_area_perf":      smallArea,
		"team_perf":            teamPerf,
		"server_time":          serverTime,
		"next_release_at":      nextReleaseAt,
		"aix_price":            aixPrice,
		"win_price":            winPrice,
		"aix_to_win_rate":      aixToWinRate,
		"exchange_fee_rate":    biz.GetExchangeFeeRate(),
		"aix_contract":         "", // TODO
		"win_contract":         s.uc.WinContract(),
		"min_usdt_recharge":    s.uc.MinUsdtRecharge(),
		"min_win_recharge":     s.uc.MinWinRecharge(),
	})
}

// HandleExchangeAixToWin 用户端：AIX → WIN 兑换
func (s *WalletService) HandleExchangeAixToWin(ctx khttp.Context) error {
	var req struct {
		Token     string `json:"token"`
		AixAmount string `json:"aix_amount"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	rec, aixLeft, winBal, err := s.uc.ExchangeAixToWin(ctx, token, req.AixAmount)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"record_id":         rec.ID,
		"from_asset":        rec.FromAsset,
		"from_amount":       rec.FromAmount,
		"to_asset":          rec.ToAsset,
		"to_amount":         rec.ToAmount,
		"exchange_price":    rec.ExchangePrice,
		"exchange_fee_rate": biz.GetExchangeFeeRate(),
		"status":            rec.Status,
		"aix_balance":       aixLeft,
		"win_balance":       winBal,
		"created_at":        rec.CreatedTime.Unix(),
	})
}

// HandleWithdrawWIN 用户端：WIN 代币提现
func (s *WalletService) HandleWithdrawWIN(ctx khttp.Context) error {
	var req struct {
		Token     string `json:"token"`
		Amount    string `json:"amount"`
		ToAddress string `json:"to_address"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	w, left, err := s.uc.CreateWinWithdraw(ctx, token, req.Amount, req.ToAddress)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"withdraw_id":  w.ID,
		"asset":        w.Asset,
		"amount":       w.Amount,
		"to_address":   w.ToAddress,
		"status":       w.Status,
		"tx_hash":      w.TxHash,
		"win_balance":  left,
		"win_contract": s.uc.WinContract(),
	})
}

// HandleCreateWinRecharge 创建 WIN 充值单（链上转账到平台收款地址后确认入账）
func (s *WalletService) HandleCreateWinRecharge(ctx khttp.Context) error {
	var req struct {
		Token  string `json:"token"`
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	recharge, err := s.uc.CreateWinRecharge(ctx, token, req.Amount)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"recharge_id":        recharge.ID,
		"asset":              biz.TokenWIN,
		"amount":             recharge.Amount,
		"deposit_address":    s.uc.DepositAddress(),
		"deposit_addresses":  s.uc.DepositAddresses(),
		"win_contract":       s.uc.WinContract(),
		"win_decimals":       s.uc.WinDecimals(),
		"token_symbol":       biz.TokenWIN,
		"message":            recharge.Message,
		"expire_at":          recharge.ExpireAt.Unix(),
		"dev_mode":           s.uc.IsDevMode(),
		"win_price":          biz.GetWinPrice(),
	})
}

// HandleConfirmWinRecharge 确认 WIN 链上充值并入账 win_balance
func (s *WalletService) HandleConfirmWinRecharge(ctx khttp.Context) error {
	var req struct {
		Token      string `json:"token"`
		RechargeID int64  `json:"recharge_id"`
		TxHash     string `json:"tx_hash"`
		Signature  string `json:"signature"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil && err != io.EOF {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid json"})
	}
	token := tokenFromRequest(ctx, req.Token)
	balance, amount, err := s.uc.ConfirmWinRecharge(ctx, token, req.RechargeID, req.TxHash, req.Signature)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"asset":       biz.TokenWIN,
		"amount":      amount,
		"win_balance": balance,
	})
}

// HandleListWinRecharges 查询本人 WIN 充值记录
func (s *WalletService) HandleListWinRecharges(ctx khttp.Context) error {
	token := tokenFromRequest(ctx, "")
	records, err := s.uc.ListWinRecharges(ctx, token)
	if err != nil {
		return err
	}
	items := make([]map[string]any, 0, len(records))
	for _, r := range records {
		item := map[string]any{
			"id": r.ID, "asset": r.Asset, "amount": r.Amount,
			"tx_hash": r.TxHash, "status": r.Status,
			"created_at": r.CreatedAt.Unix(),
		}
		if r.ConfirmedAt != nil {
			item["confirmed_at"] = r.ConfirmedAt.Unix()
		}
		items = append(items, item)
	}
	return ctx.JSON(http.StatusOK, map[string]any{"recharges": items})
}

// HandleExchangeRecords 用户端：查询本人的 AIX→WIN 兑换记录
func (s *WalletService) HandleExchangeRecords(ctx khttp.Context) error {
	token := tokenFromRequest(ctx, "")
	records, err := s.uc.ListExchangeRecords(ctx, token)
	if err != nil {
		return err
	}
	items := make([]map[string]any, 0, len(records))
	for _, r := range records {
		items = append(items, map[string]any{
			"id":             r.ID,
			"from_asset":     r.FromAsset,
			"from_amount":    r.FromAmount,
			"to_asset":       r.ToAsset,
			"to_amount":      r.ToAmount,
			"fee_amount":     r.FeeAmount,
			"fee_rate":       r.FeeRate,
			"exchange_price": r.ExchangePrice,
			"status":         r.Status,
			"remark":         r.Remark,
			"created_at":     r.CreatedTime.Unix(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]any{"records": items})
}
