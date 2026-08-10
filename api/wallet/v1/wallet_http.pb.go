package v1

import (
	context "context"

	http "github.com/go-kratos/kratos/v2/transport/http"
)

const _ = http.SupportPackageIsVersion1

type WalletHTTPServer interface {
	GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceReply, error)
	CreateRecharge(context.Context, *CreateRechargeRequest) (*CreateRechargeReply, error)
	ConfirmRecharge(context.Context, *ConfirmRechargeRequest) (*ConfirmRechargeReply, error)
	ListRecharges(context.Context, *ListRechargesRequest) (*ListRechargesReply, error)
	CreateWithdraw(context.Context, *CreateWithdrawRequest) (*CreateWithdrawReply, error)
	ClaimToAccount(context.Context, *ClaimToAccountRequest) (*ClaimToAccountReply, error)
	ListWithdrawals(context.Context, *ListWithdrawalsRequest) (*ListWithdrawalsReply, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error)
	Subscribe(context.Context, *SubscribeRequest) (*SubscribeReply, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersReply, error)
	ListReleaseRecords(context.Context, *ListReleaseRecordsRequest) (*ListReleaseRecordsReply, error)
	ListReferralRewards(context.Context, *ListReferralRewardsRequest) (*ListReferralRewardsReply, error)
	ListEcoRewards(context.Context, *ListEcoRewardsRequest) (*ListEcoRewardsReply, error)
	ListClaimRecords(context.Context, *ListClaimRecordsRequest) (*ListClaimRecordsReply, error)
}

func RegisterWalletHTTPServer(s *http.Server, srv WalletHTTPServer) {
	r := s.Route("/")
	r.GET("/v1/wallet/balance", _Wallet_GetBalance0_HTTP_Handler(srv))
	r.POST("/v1/wallet/recharge", _Wallet_CreateRecharge0_HTTP_Handler(srv))
	r.POST("/v1/wallet/recharge/confirm", _Wallet_ConfirmRecharge0_HTTP_Handler(srv))
	r.GET("/v1/wallet/recharges", _Wallet_ListRecharges0_HTTP_Handler(srv))
	r.POST("/v1/wallet/withdraw", _Wallet_CreateWithdraw0_HTTP_Handler(srv))
	r.POST("/v1/wallet/claim", _Wallet_ClaimToAccount0_HTTP_Handler(srv))
	r.GET("/v1/wallet/withdrawals", _Wallet_ListWithdrawals0_HTTP_Handler(srv))
	r.GET("/v1/wallet/products", _Wallet_ListProducts0_HTTP_Handler(srv))
	r.POST("/v1/wallet/subscribe", _Wallet_Subscribe0_HTTP_Handler(srv))
	r.GET("/v1/wallet/orders", _Wallet_ListOrders0_HTTP_Handler(srv))
	r.GET("/v1/wallet/releases", _Wallet_ListReleaseRecords0_HTTP_Handler(srv))
	r.GET("/v1/wallet/referral-rewards", _Wallet_ListReferralRewards0_HTTP_Handler(srv))
	r.GET("/v1/wallet/eco-rewards", _Wallet_ListEcoRewards0_HTTP_Handler(srv))
	r.GET("/v1/wallet/claims", _Wallet_ListClaimRecords0_HTTP_Handler(srv))
}

func _Wallet_GetBalance0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetBalanceRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/GetBalance")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetBalance(ctx, req.(*GetBalanceRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*GetBalanceReply))
	}
}

func _Wallet_CreateRecharge0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateRechargeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/CreateRecharge")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateRecharge(ctx, req.(*CreateRechargeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*CreateRechargeReply))
	}
}

func _Wallet_ConfirmRecharge0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ConfirmRechargeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ConfirmRecharge")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ConfirmRecharge(ctx, req.(*ConfirmRechargeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ConfirmRechargeReply))
	}
}

func _Wallet_ListRecharges0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListRechargesRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListRecharges")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListRecharges(ctx, req.(*ListRechargesRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListRechargesReply))
	}
}

func _Wallet_CreateWithdraw0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateWithdrawRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/CreateWithdraw")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateWithdraw(ctx, req.(*CreateWithdrawRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*CreateWithdrawReply))
	}
}

func _Wallet_ClaimToAccount0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ClaimToAccountRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ClaimToAccount")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ClaimToAccount(ctx, req.(*ClaimToAccountRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ClaimToAccountReply))
	}
}

func _Wallet_ListWithdrawals0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListWithdrawalsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListWithdrawals")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListWithdrawals(ctx, req.(*ListWithdrawalsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListWithdrawalsReply))
	}
}

func _Wallet_ListProducts0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListProductsRequest
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListProducts")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListProducts(ctx, req.(*ListProductsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListProductsReply))
	}
}

func _Wallet_Subscribe0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in SubscribeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/Subscribe")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Subscribe(ctx, req.(*SubscribeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*SubscribeReply))
	}
}

func _Wallet_ListOrders0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListOrdersRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListOrders")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListOrders(ctx, req.(*ListOrdersRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListOrdersReply))
	}
}

func _Wallet_ListReleaseRecords0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListReleaseRecordsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListReleaseRecords")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListReleaseRecords(ctx, req.(*ListReleaseRecordsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListReleaseRecordsReply))
	}
}

func _Wallet_ListReferralRewards0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListReferralRewardsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListReferralRewards")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListReferralRewards(ctx, req.(*ListReferralRewardsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListReferralRewardsReply))
	}
}

func _Wallet_ListEcoRewards0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListEcoRewardsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListEcoRewards")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListEcoRewards(ctx, req.(*ListEcoRewardsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListEcoRewardsReply))
	}
}

func _Wallet_ListClaimRecords0_HTTP_Handler(srv WalletHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListClaimRecordsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/wallet.v1.Wallet/ListClaimRecords")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListClaimRecords(ctx, req.(*ListClaimRecordsRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListClaimRecordsReply))
	}
}
