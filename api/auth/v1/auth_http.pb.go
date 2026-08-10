package v1

import (
	context "context"

	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

const _ = http.SupportPackageIsVersion1

type AuthHTTPServer interface {
	GetChallenge(context.Context, *GetChallengeRequest) (*GetChallengeReply, error)
	Login(context.Context, *LoginRequest) (*LoginReply, error)
	GetProfile(context.Context, *GetProfileRequest) (*GetProfileReply, error)
	ListInvitees(context.Context, *ListInviteesRequest) (*ListInviteesReply, error)
}

func RegisterAuthHTTPServer(s *http.Server, srv AuthHTTPServer) {
	r := s.Route("/")
	r.GET("/v1/auth/challenge", _Auth_GetChallenge0_HTTP_Handler(srv))
	r.POST("/v1/auth/login", _Auth_Login0_HTTP_Handler(srv))
	r.GET("/v1/auth/profile", _Auth_GetProfile0_HTTP_Handler(srv))
	r.GET("/v1/auth/invitees", _Auth_ListInvitees0_HTTP_Handler(srv))
}

func _Auth_GetChallenge0_HTTP_Handler(srv AuthHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetChallengeRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/auth.v1.Auth/GetChallenge")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetChallenge(ctx, req.(*GetChallengeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*GetChallengeReply))
	}
}

func _Auth_Login0_HTTP_Handler(srv AuthHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in LoginRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/auth.v1.Auth/Login")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Login(ctx, req.(*LoginRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*LoginReply))
	}
}

func _Auth_GetProfile0_HTTP_Handler(srv AuthHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetProfileRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/auth.v1.Auth/GetProfile")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetProfile(ctx, req.(*GetProfileRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*GetProfileReply))
	}
}

func _Auth_ListInvitees0_HTTP_Handler(srv AuthHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListInviteesRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/auth.v1.Auth/ListInvitees")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListInvitees(ctx, req.(*ListInviteesRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out.(*ListInviteesReply))
	}
}

type AuthHTTPClient interface {
	GetChallenge(ctx context.Context, req *GetChallengeRequest, opts ...http.CallOption) (*GetChallengeReply, error)
	Login(ctx context.Context, req *LoginRequest, opts ...http.CallOption) (*LoginReply, error)
	GetProfile(ctx context.Context, req *GetProfileRequest, opts ...http.CallOption) (*GetProfileReply, error)
	ListInvitees(ctx context.Context, req *ListInviteesRequest, opts ...http.CallOption) (*ListInviteesReply, error)
}

type AuthHTTPClientImpl struct {
	cc *http.Client
}

func NewAuthHTTPClient(client *http.Client) AuthHTTPClient {
	return &AuthHTTPClientImpl{client}
}

func (c *AuthHTTPClientImpl) GetChallenge(ctx context.Context, in *GetChallengeRequest, opts ...http.CallOption) (*GetChallengeReply, error) {
	var out GetChallengeReply
	pattern := "/v1/auth/challenge"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation("/auth.v1.Auth/GetChallenge"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	return &out, err
}

func (c *AuthHTTPClientImpl) Login(ctx context.Context, in *LoginRequest, opts ...http.CallOption) (*LoginReply, error) {
	var out LoginReply
	opts = append(opts, http.Operation("/auth.v1.Auth/Login"))
	err := c.cc.Invoke(ctx, "POST", "/v1/auth/login", in, &out, opts...)
	return &out, err
}

func (c *AuthHTTPClientImpl) GetProfile(ctx context.Context, in *GetProfileRequest, opts ...http.CallOption) (*GetProfileReply, error) {
	var out GetProfileReply
	pattern := "/v1/auth/profile"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation("/auth.v1.Auth/GetProfile"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	return &out, err
}

func (c *AuthHTTPClientImpl) ListInvitees(ctx context.Context, in *ListInviteesRequest, opts ...http.CallOption) (*ListInviteesReply, error) {
	var out ListInviteesReply
	pattern := "/v1/auth/invitees"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation("/auth.v1.Auth/ListInvitees"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	return &out, err
}
