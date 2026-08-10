package v1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion7

type AuthClient interface {
	GetChallenge(ctx context.Context, in *GetChallengeRequest, opts ...grpc.CallOption) (*GetChallengeReply, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginReply, error)
	GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*GetProfileReply, error)
	ListInvitees(ctx context.Context, in *ListInviteesRequest, opts ...grpc.CallOption) (*ListInviteesReply, error)
}

type authClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthClient(cc grpc.ClientConnInterface) AuthClient {
	return &authClient{cc}
}

func (c *authClient) GetChallenge(ctx context.Context, in *GetChallengeRequest, opts ...grpc.CallOption) (*GetChallengeReply, error) {
	out := new(GetChallengeReply)
	err := c.cc.Invoke(ctx, "/auth.v1.Auth/GetChallenge", in, out, opts...)
	return out, err
}

func (c *authClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginReply, error) {
	out := new(LoginReply)
	err := c.cc.Invoke(ctx, "/auth.v1.Auth/Login", in, out, opts...)
	return out, err
}

func (c *authClient) GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*GetProfileReply, error) {
	out := new(GetProfileReply)
	err := c.cc.Invoke(ctx, "/auth.v1.Auth/GetProfile", in, out, opts...)
	return out, err
}

func (c *authClient) ListInvitees(ctx context.Context, in *ListInviteesRequest, opts ...grpc.CallOption) (*ListInviteesReply, error) {
	out := new(ListInviteesReply)
	err := c.cc.Invoke(ctx, "/auth.v1.Auth/ListInvitees", in, out, opts...)
	return out, err
}

type AuthServer interface {
	GetChallenge(context.Context, *GetChallengeRequest) (*GetChallengeReply, error)
	Login(context.Context, *LoginRequest) (*LoginReply, error)
	GetProfile(context.Context, *GetProfileRequest) (*GetProfileReply, error)
	ListInvitees(context.Context, *ListInviteesRequest) (*ListInviteesReply, error)
	mustEmbedUnimplementedAuthServer()
}

type UnimplementedAuthServer struct{}

func (UnimplementedAuthServer) GetChallenge(context.Context, *GetChallengeRequest) (*GetChallengeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChallenge not implemented")
}
func (UnimplementedAuthServer) Login(context.Context, *LoginRequest) (*LoginReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAuthServer) GetProfile(context.Context, *GetProfileRequest) (*GetProfileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProfile not implemented")
}
func (UnimplementedAuthServer) ListInvitees(context.Context, *ListInviteesRequest) (*ListInviteesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInvitees not implemented")
}
func (UnimplementedAuthServer) mustEmbedUnimplementedAuthServer() {}

func RegisterAuthServer(s grpc.ServiceRegistrar, srv AuthServer) {
	s.RegisterService(&Auth_ServiceDesc, srv)
}

func _Auth_GetChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetChallengeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).GetChallenge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.Auth/GetChallenge"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).GetChallenge(ctx, req.(*GetChallengeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.Auth/Login"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_GetProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).GetProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.Auth/GetProfile"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).GetProfile(ctx, req.(*GetProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_ListInvitees_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInviteesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).ListInvitees(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.Auth/ListInvitees"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).ListInvitees(ctx, req.(*ListInviteesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Auth_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.v1.Auth",
	HandlerType: (*AuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetChallenge", Handler: _Auth_GetChallenge_Handler},
		{MethodName: "Login", Handler: _Auth_Login_Handler},
		{MethodName: "GetProfile", Handler: _Auth_GetProfile_Handler},
		{MethodName: "ListInvitees", Handler: _Auth_ListInvitees_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth/v1/auth.proto",
}
