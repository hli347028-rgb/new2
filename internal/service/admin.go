package service

import (
	"context"

	v1 "backend/api/admin/v1"
	"backend/internal/biz"
	"backend/internal/conf"
)

type AdminService struct {
	uc *biz.AdminUsecase
}

func NewAdminService(uc *biz.AdminUsecase) *AdminService {
	return &AdminService{uc: uc}
}

func (s *AdminService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersReply, error) {
	users, err := s.uc.ListUsers(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AdminUser, 0, len(users))
	for _, u := range users {
		items = append(items, toAdminUser(u))
	}
	return &v1.ListUsersReply{Users: items}, nil
}

func (s *AdminService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserReply, error) {
	updated, err := s.uc.UpdateUser(ctx, resolveToken(ctx, req.Token), &biz.AdminUserUpdate{
		UserID:            req.UserID,
		Balance:           req.Balance,
		ReleasedBalance:   req.ReleasedBalance,
		Role:              req.Role,
		CommunityLevel:    req.CommunityLevel,
		SetCommunityLevel: req.CommunityLevel != "",
		CommunityStake:    req.CommunityStake,
		TeamStake:         req.TeamStake,
		InviterID:         req.InviterID,
		WithdrawReset:     req.WithdrawReset,
		WinBalance:        req.WinBalance,
		PendingMgmtReward: req.PendingMgmtReward,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateUserReply{User: toAdminUser(updated)}, nil
}

func (s *AdminService) GetConfig(ctx context.Context, req *v1.GetConfigRequest) (*v1.GetConfigReply, error) {
	cfg, err := s.uc.GetSystemConfig(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	return &v1.GetConfigReply{Config: toSystemConfig(cfg)}, nil
}

func (s *AdminService) UpdateConfig(ctx context.Context, req *v1.UpdateConfigRequest) (*v1.UpdateConfigReply, error) {
	cfg, err := s.uc.UpdateSystemConfig(ctx, resolveToken(ctx, req.Token), &conf.SystemConfigSnapshot{
		JwtSecret:       req.JwtSecret,
		ChallengeTTL:    req.ChallengeTTL,
		AdminAddresses:  req.AdminAddresses,
		DepositAddress:  req.DepositAddress,
		UsdtContract:    req.UsdtContract,
		UsdtDecimals:    req.UsdtDecimals,
		RPCURL:          req.RPCURL,
		MinSubscribe:    req.MinSubscribe,
		WithdrawFeeRate: req.WithdrawFeeRate,
		WinPrice:        req.WinPrice,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateConfigReply{Config: toSystemConfig(cfg)}, nil
}

func (s *AdminService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsReply, error) {
	products, err := s.uc.ListAllProducts(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AdminProduct, 0, len(products))
	for _, p := range products {
		items = append(items, &v1.AdminProduct{
			ID: p.ID, Name: p.Name, Description: p.Description,
			Price: p.Price, Stock: p.Stock, Status: p.Status,
		})
	}
	return &v1.ListProductsReply{Products: items}, nil
}

func (s *AdminService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductReply, error) {
	product, err := s.uc.UpdateProduct(ctx, resolveToken(ctx, req.Token), &biz.Product{
		ID: req.ProductID, Name: req.Name, Description: req.Description,
		Price: req.Price, Stock: req.Stock, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateProductReply{Product: &v1.AdminProduct{
		ID: product.ID, Name: product.Name, Description: product.Description,
		Price: product.Price, Stock: product.Stock, Status: product.Status,
	}}, nil
}

func (s *AdminService) TriggerSettlement(ctx context.Context, req *v1.TriggerSettlementRequest) (*v1.TriggerSettlementReply, error) {
	if err := s.uc.TriggerSettlement(ctx, resolveToken(ctx, req.Token), req.SettlementDate); err != nil {
		return nil, err
	}
	return &v1.TriggerSettlementReply{Message: "结算任务已触发"}, nil
}

func (s *AdminService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersReply, error) {
	orders, err := s.uc.ListOrders(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AdminOrder, 0, len(orders))
	for _, o := range orders {
		items = append(items, toAdminOrder(o))
	}
	return &v1.ListOrdersReply{Orders: items}, nil
}

func (s *AdminService) UpdateOrder(ctx context.Context, req *v1.UpdateOrderRequest) (*v1.UpdateOrderReply, error) {
	updated, err := s.uc.UpdateOrder(ctx, resolveToken(ctx, req.Token), &biz.AdminOrderUpdate{
		OrderID:        req.OrderID,
		Quantity:       req.Quantity,
		TotalAmount:    req.TotalAmount,
		Status:         req.Status,
		ExitMultiplier: req.ExitMultiplier,
		ExitTarget:     req.ExitTarget,
		ReleasedAmount: req.ReleasedAmount,
		ReleaseDay:     req.ReleaseDay,
		CycleDay:       req.CycleDay,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateOrderReply{Order: toAdminOrder(updated)}, nil
}

func toAdminUser(u *biz.AdminUserDetail) *v1.AdminUser {
	if u == nil || u.User == nil {
		return nil
	}
	return &v1.AdminUser{
		ID: u.User.ID, Address: u.User.Address, Balance: u.User.Balance,
		ReleasedBalance: u.User.ReleasedBalance,
		Role:            u.User.Role, InviterID: u.User.InviterID, InviterAddress: u.User.InviterAddress,
		InviteeCount: u.InviteeCount, CommunityLevel: u.User.CommunityLevel,
		CommunityStake: u.User.CommunityStake, TeamStake: u.User.TeamStake,
		LargeAreaPerf:    u.User.LargeAreaPerf,
		ShareProfitTotal: u.User.ShareProfitTotal, EcoRewardTotal: u.User.EcoRewardTotal,
		WinBalance:        u.User.WinBalance,
		PendingMgmtReward: u.User.PendingMgmtReward,
		WithdrawReset:     u.WithdrawReset, CreatedAt: u.User.CreatedAt.Unix(),
	}
}

func toSystemConfig(cfg *conf.SystemConfigSnapshot) *v1.SystemConfig {
	if cfg == nil {
		return nil
	}
	return &v1.SystemConfig{
		JwtSecret: cfg.JwtSecret, ChallengeTTL: cfg.ChallengeTTL,
		AdminAddresses: cfg.AdminAddresses, DepositAddress: cfg.DepositAddress,
		UsdtContract: cfg.UsdtContract, UsdtDecimals: cfg.UsdtDecimals,
		RPCURL: cfg.RPCURL, MinSubscribe: cfg.MinSubscribe,
		WithdrawFeeRate: cfg.WithdrawFeeRate,
		WinPrice:        cfg.WinPrice,
	}
}

func toAdminOrder(o *biz.AdminOrderDetail) *v1.AdminOrder {
	if o == nil || o.Order == nil {
		return nil
	}
	return &v1.AdminOrder{
		ID: o.Order.ID, UserID: o.Order.UserID, UserAddress: o.UserAddress,
		ProductID: o.Order.ProductID, ProductName: o.Order.ProductName,
		Quantity: o.Order.Quantity, TotalAmount: o.Order.TotalAmount,
		Status: o.Order.Status, ExitMultiplier: o.Order.ExitMultiplier,
		ExitTarget: o.Order.ExitTarget, ReleasedAmount: o.Order.ReleasedAmount,
		ReleaseDay: o.Order.ReleaseDay, CycleDay: o.Order.CycleDay,
		CreatedAt: o.Order.CreatedAt.Unix(),
	}
}
