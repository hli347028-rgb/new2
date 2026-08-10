package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	authmw "backend/internal/middleware"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterWalletExtraRoutes mounts AIX-specific routes without regenerating protos.
func RegisterWalletExtraRoutes(srv *khttp.Server, wallet *WalletService) {
	r := srv.Route("/")
	r.POST("/v1/wallet/subscribe-aix", wallet.HandleSubscribeAIX)
	r.POST("/v1/wallet/transfer", wallet.HandleTransfer)
	r.POST("/v1/wallet/recharge-to-reward", wallet.HandleRechargeToReward)
	r.POST("/v1/wallet/withdraw-aix", wallet.HandleWithdrawAIX)
	r.GET("/v1/wallet/aix-price", wallet.HandleAixPrice)
	r.GET("/v1/wallet/rewards", wallet.HandleRewards)
	r.GET("/v1/wallet/aix-profile", wallet.HandleAixProfile)
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
	return ctx.JSON(http.StatusOK, map[string]any{
		"order_id":     order.ID,
		"principal":    order.Principal,
		"exit_cap":     order.ExitCap,
		"fund_source":  order.FundSource,
		"direct_base":  order.DirectBase,
		"status":       order.Status,
		"balance":      bal,
		"total_amount": order.Principal,
	})
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
		"withdraw_id": w.ID,
		"asset":       "AIX",
		"amount":      w.Amount,
		"to_address":  w.ToAddress,
		"status":      w.Status,
		"tx_hash":     w.TxHash, // 合约未就绪前为空
		"aix_balance": left,
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
	return ctx.JSON(http.StatusOK, map[string]any{
		"address":           user.Address,
		"usdt_recharge":     recharge,
		"usdt_reward":       reward,
		"aix_balance":       aix,
		"static_usdt_total": staticTotal,
		"pending_amount":    pending,
		"unexited_amount":   unexited,
		"total_nodes":       totalNodes,
		"mgmt_level":        mgmtLevel,
		"large_area_perf":   largeArea,
		"small_area_perf":   smallArea,
		"team_perf":         teamPerf,
		"server_time":       serverTime,
		"next_release_at":   nextReleaseAt,
		"aix_contract":      "", // TODO
	})
}
