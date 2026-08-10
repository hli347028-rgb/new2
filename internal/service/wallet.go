package service

import (
	"context"

	v1 "backend/api/wallet/v1"
	"backend/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
)

var errAIXUnsupported = errors.BadRequest("NOT_SUPPORTED", "not supported in AIX")

type WalletService struct {
	v1.UnimplementedWalletServer
	uc *biz.WalletUsecase
}

func NewWalletService(uc *biz.WalletUsecase) *WalletService {
	return &WalletService{uc: uc}
}

func (s *WalletService) GetBalance(ctx context.Context, req *v1.GetBalanceRequest) (*v1.GetBalanceReply, error) {
	user, summary, recharge, reward, pending, aix, unexited, nextReleaseAt, serverTime, err := s.uc.GetBalance(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	totalNodes := int32(0)
	if summary != nil {
		totalNodes = summary.TotalNodes
	}
	// Mapping: balance=usdt_recharge, released=usdt_reward, claimed=aix_balance
	return &v1.GetBalanceReply{
		Address:         user.Address,
		Balance:         recharge,
		ReleasedBalance: reward,
		ClaimableAmount: aix,
		PendingAmount:   pending,
		ClaimedAmount:   aix,
		TotalNodes:      totalNodes,
		ServerTime:      serverTime,
		NextReleaseAt:   nextReleaseAt,
		UnexitedAmount:  unexited,
	}, nil
}

func (s *WalletService) CreateRecharge(ctx context.Context, req *v1.CreateRechargeRequest) (*v1.CreateRechargeReply, error) {
	recharge, err := s.uc.CreateRecharge(ctx, resolveToken(ctx, req.Token), req.Amount)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRechargeReply{
		RechargeID:       recharge.ID,
		Amount:           recharge.Amount,
		DepositAddress:   s.uc.DepositAddress(),
		DepositAddresses: s.uc.DepositAddresses(),
		SplitAmounts:     s.uc.SplitDepositAmounts(recharge.Amount),
		UsdtContract:     s.uc.UsdtContract(),
		TokenSymbol:      biz.TokenUSDT,
		Message:          recharge.Message,
		ExpireAt:         recharge.ExpireAt.Unix(),
		DevMode:          s.uc.IsDevMode(),
	}, nil
}

func (s *WalletService) ConfirmRecharge(ctx context.Context, req *v1.ConfirmRechargeRequest) (*v1.ConfirmRechargeReply, error) {
	balance, amount, err := s.uc.ConfirmRecharge(ctx, resolveToken(ctx, req.Token), req.RechargeID, req.TxHash, req.TxHashes, req.Signature)
	if err != nil {
		return nil, err
	}
	return &v1.ConfirmRechargeReply{Balance: balance, Amount: amount}, nil
}

func (s *WalletService) ListRecharges(ctx context.Context, req *v1.ListRechargesRequest) (*v1.ListRechargesReply, error) {
	records, err := s.uc.ListRecharges(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.RechargeRecord, 0, len(records))
	for _, r := range records {
		item := &v1.RechargeRecord{
			ID: r.ID, Amount: r.Amount, TxHash: r.TxHash, Status: r.Status, CreatedAt: r.CreatedAt.Unix(),
		}
		if r.ConfirmedAt != nil {
			item.ConfirmedAt = r.ConfirmedAt.Unix()
		}
		items = append(items, item)
	}
	return &v1.ListRechargesReply{Recharges: items}, nil
}

func (s *WalletService) CreateWithdraw(ctx context.Context, req *v1.CreateWithdrawRequest) (*v1.CreateWithdrawReply, error) {
	return nil, errAIXUnsupported
}

func (s *WalletService) ClaimToAccount(ctx context.Context, req *v1.ClaimToAccountRequest) (*v1.ClaimToAccountReply, error) {
	return nil, errAIXUnsupported
}

func (s *WalletService) ListWithdrawals(ctx context.Context, req *v1.ListWithdrawalsRequest) (*v1.ListWithdrawalsReply, error) {
	list, err := s.uc.ListWithdrawals(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.WithdrawalRecord, 0, len(list))
	for _, w := range list {
		items = append(items, &v1.WithdrawalRecord{
			ID: w.ID, Amount: w.Amount, Fee: w.Fee, NetAmount: w.NetAmount,
			ToAddress: w.ToAddress, Status: w.Status, TxHash: w.TxHash, CreatedAt: w.CreatedAt.Unix(),
		})
	}
	return &v1.ListWithdrawalsReply{Withdrawals: items}, nil
}

func (s *WalletService) ListProducts(ctx context.Context, _ *v1.ListProductsRequest) (*v1.ListProductsReply, error) {
	return nil, errAIXUnsupported
}

func (s *WalletService) Subscribe(ctx context.Context, req *v1.SubscribeRequest) (*v1.SubscribeReply, error) {
	// Proto has no pay_from; accept header Pay-From, else use subscribe-aix JSON route.
	payFrom := ""
	if tr, ok := transport.FromServerContext(ctx); ok {
		payFrom = tr.RequestHeader().Get("Pay-From")
		if payFrom == "" {
			payFrom = tr.RequestHeader().Get("pay_from")
		}
	}
	if req.Amount != "" && payFrom != "" {
		order, bal, err := s.uc.SubscribeAIX(ctx, resolveToken(ctx, req.Token), req.Amount, payFrom)
		if err != nil {
			return nil, err
		}
		return &v1.SubscribeReply{
			OrderID: order.ID, ProductName: order.FundSource,
			TotalAmount: order.Principal, Balance: bal,
		}, nil
	}
	return nil, errors.BadRequest("PAY_FROM_REQUIRED", "use POST /v1/wallet/subscribe-aix with amount and pay_from=recharge|reward")
}

func (s *WalletService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersReply, error) {
	orders, err := s.uc.ListOrders(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.Order, 0, len(orders))
	for _, o := range orders {
		items = append(items, &v1.Order{
			ID:             o.ID,
			ProductID:      0,
			ProductName:    o.FundSource,
			Quantity:       1,
			TotalAmount:    o.Principal,
			CreatedAt:      o.CreatedTime.Unix(),
			Status:         o.Status,
			ExitMultiplier: o.ExitMultiplier,
			ExitTarget:     o.ExitCap,
			ReleasedAmount: o.EarnedTotal,
			ReleaseDay:     0,
			CycleDay:       0,
		})
	}
	return &v1.ListOrdersReply{Orders: items}, nil
}

func (s *WalletService) ListReleaseRecords(ctx context.Context, req *v1.ListReleaseRecordsRequest) (*v1.ListReleaseRecordsReply, error) {
	records, err := s.uc.ListReleaseRecords(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ReleaseRecord, 0, len(records))
	for _, r := range records {
		items = append(items, &v1.ReleaseRecord{
			ID: r.ID, OrderID: r.OrderID, SettlementDate: r.SettlementDate,
			Rate: r.Rate, Amount: r.Amount, CreatedAt: r.CreatedAt.Unix(),
		})
	}
	return &v1.ListReleaseRecordsReply{Records: items}, nil
}

func (s *WalletService) ListReferralRewards(ctx context.Context, req *v1.ListReferralRewardsRequest) (*v1.ListReferralRewardsReply, error) {
	rewards, err := s.uc.ListReferralRewards(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ReferralReward, 0, len(rewards))
	for _, r := range rewards {
		sourceAddr := ""
		if src, err := s.uc.FindUserAddress(ctx, r.SourceUserID); err == nil {
			sourceAddr = src
		}
		items = append(items, &v1.ReferralReward{
			ID: r.ID, SourceUserID: r.SourceUserID, SourceAddress: sourceAddr,
			SourceOrderID: r.SourceOrderID, Generation: r.Generation,
			BaseAmount: r.BaseAmount, Rate: r.Rate, RewardAmount: r.RewardAmount,
			SettlementDate: r.SettlementDate, CreatedAt: r.CreatedAt.Unix(),
		})
	}
	return &v1.ListReferralRewardsReply{Rewards: items}, nil
}

func (s *WalletService) ListEcoRewards(ctx context.Context, req *v1.ListEcoRewardsRequest) (*v1.ListEcoRewardsReply, error) {
	return nil, errAIXUnsupported
}

func (s *WalletService) ListClaimRecords(ctx context.Context, req *v1.ListClaimRecordsRequest) (*v1.ListClaimRecordsReply, error) {
	return nil, errAIXUnsupported
}

// Usecase exposes usecase for custom HTTP routes.
func (s *WalletService) Usecase() *biz.WalletUsecase { return s.uc }
