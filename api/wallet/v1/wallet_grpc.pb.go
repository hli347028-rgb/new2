package v1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion7

type WalletServer interface {
	GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceReply, error)
	CreateRecharge(context.Context, *CreateRechargeRequest) (*CreateRechargeReply, error)
	ConfirmRecharge(context.Context, *ConfirmRechargeRequest) (*ConfirmRechargeReply, error)
	ListRecharges(context.Context, *ListRechargesRequest) (*ListRechargesReply, error)
	CreateWithdraw(context.Context, *CreateWithdrawRequest) (*CreateWithdrawReply, error)
	ListWithdrawals(context.Context, *ListWithdrawalsRequest) (*ListWithdrawalsReply, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error)
	Subscribe(context.Context, *SubscribeRequest) (*SubscribeReply, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersReply, error)
	ListReleaseRecords(context.Context, *ListReleaseRecordsRequest) (*ListReleaseRecordsReply, error)
	ListReferralRewards(context.Context, *ListReferralRewardsRequest) (*ListReferralRewardsReply, error)
	ListEcoRewards(context.Context, *ListEcoRewardsRequest) (*ListEcoRewardsReply, error)
	mustEmbedUnimplementedWalletServer()
}

type UnimplementedWalletServer struct{}

func (UnimplementedWalletServer) GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (UnimplementedWalletServer) CreateRecharge(context.Context, *CreateRechargeRequest) (*CreateRechargeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRecharge not implemented")
}
func (UnimplementedWalletServer) ConfirmRecharge(context.Context, *ConfirmRechargeRequest) (*ConfirmRechargeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmRecharge not implemented")
}
func (UnimplementedWalletServer) ListRecharges(context.Context, *ListRechargesRequest) (*ListRechargesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRecharges not implemented")
}
func (UnimplementedWalletServer) CreateWithdraw(context.Context, *CreateWithdrawRequest) (*CreateWithdrawReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWithdraw not implemented")
}
func (UnimplementedWalletServer) ListWithdrawals(context.Context, *ListWithdrawalsRequest) (*ListWithdrawalsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWithdrawals not implemented")
}
func (UnimplementedWalletServer) ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProducts not implemented")
}
func (UnimplementedWalletServer) Subscribe(context.Context, *SubscribeRequest) (*SubscribeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedWalletServer) ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}
func (UnimplementedWalletServer) ListReleaseRecords(context.Context, *ListReleaseRecordsRequest) (*ListReleaseRecordsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReleaseRecords not implemented")
}
func (UnimplementedWalletServer) ListReferralRewards(context.Context, *ListReferralRewardsRequest) (*ListReferralRewardsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReferralRewards not implemented")
}
func (UnimplementedWalletServer) ListEcoRewards(context.Context, *ListEcoRewardsRequest) (*ListEcoRewardsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListEcoRewards not implemented")
}
func (UnimplementedWalletServer) mustEmbedUnimplementedWalletServer() {}

func RegisterWalletServer(s grpc.ServiceRegistrar, srv WalletServer) {
	s.RegisterService(&Wallet_ServiceDesc, srv)
}

var Wallet_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wallet.v1.Wallet",
	HandlerType: (*WalletServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetBalance", Handler: _Wallet_GetBalance_Handler},
		{MethodName: "CreateRecharge", Handler: _Wallet_CreateRecharge_Handler},
		{MethodName: "ConfirmRecharge", Handler: _Wallet_ConfirmRecharge_Handler},
		{MethodName: "ListRecharges", Handler: _Wallet_ListRecharges_Handler},
		{MethodName: "CreateWithdraw", Handler: _Wallet_CreateWithdraw_Handler},
		{MethodName: "ListWithdrawals", Handler: _Wallet_ListWithdrawals_Handler},
		{MethodName: "ListProducts", Handler: _Wallet_ListProducts_Handler},
		{MethodName: "Subscribe", Handler: _Wallet_Subscribe_Handler},
		{MethodName: "ListOrders", Handler: _Wallet_ListOrders_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "wallet/v1/wallet.proto",
}

func _Wallet_GetBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).GetBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/GetBalance"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).GetBalance(ctx, req.(*GetBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_CreateRecharge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRechargeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).CreateRecharge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/CreateRecharge"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).CreateRecharge(ctx, req.(*CreateRechargeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_ConfirmRecharge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfirmRechargeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).ConfirmRecharge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/ConfirmRecharge"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).ConfirmRecharge(ctx, req.(*ConfirmRechargeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_ListRecharges_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRechargesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).ListRecharges(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/ListRecharges"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).ListRecharges(ctx, req.(*ListRechargesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_CreateWithdraw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWithdrawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).CreateWithdraw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/CreateWithdraw"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).CreateWithdraw(ctx, req.(*CreateWithdrawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_ListWithdrawals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListWithdrawalsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).ListWithdrawals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/ListWithdrawals"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).ListWithdrawals(ctx, req.(*ListWithdrawalsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_ListProducts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProductsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).ListProducts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/ListProducts"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).ListProducts(ctx, req.(*ListProductsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_Subscribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscribeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).Subscribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/Subscribe"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).Subscribe(ctx, req.(*SubscribeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Wallet_ListOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServer).ListOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/wallet.v1.Wallet/ListOrders"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServer).ListOrders(ctx, req.(*ListOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}
