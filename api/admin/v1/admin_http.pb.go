package v1

import (
	context "context"

	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

const _ = http.SupportPackageIsVersion1

type AdminHTTPServer interface {
	ListUsers(context.Context, *ListUsersRequest) (*ListUsersReply, error)
	UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserReply, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigReply, error)
	UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigReply, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsReply, error)
	UpdateProduct(context.Context, *UpdateProductRequest) (*UpdateProductReply, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersReply, error)
	UpdateOrder(context.Context, *UpdateOrderRequest) (*UpdateOrderReply, error)
	TriggerSettlement(context.Context, *TriggerSettlementRequest) (*TriggerSettlementReply, error)
}

func RegisterAdminHTTPServer(s *http.Server, srv AdminHTTPServer) {
	r := s.Route("/")
	r.GET("/v1/admin/users", _Admin_ListUsers0_HTTP_Handler(srv))
	r.PUT("/v1/admin/users", _Admin_UpdateUser0_HTTP_Handler(srv))
	r.GET("/v1/admin/config", _Admin_GetConfig0_HTTP_Handler(srv))
	r.PUT("/v1/admin/config", _Admin_UpdateConfig0_HTTP_Handler(srv))
	r.GET("/v1/admin/products", _Admin_ListProducts0_HTTP_Handler(srv))
	r.PUT("/v1/admin/products", _Admin_UpdateProduct0_HTTP_Handler(srv))
	r.GET("/v1/admin/orders", _Admin_ListOrders0_HTTP_Handler(srv))
	r.PUT("/v1/admin/orders", _Admin_UpdateOrder0_HTTP_Handler(srv))
	r.POST("/v1/admin/settlement/trigger", _Admin_TriggerSettlement0_HTTP_Handler(srv))
}

func _Admin_ListUsers0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListUsersRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		out, err := srv.ListUsers(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_UpdateUser0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateUserRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		out, err := srv.UpdateUser(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_GetConfig0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetConfigRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		out, err := srv.GetConfig(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_UpdateConfig0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateConfigRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		out, err := srv.UpdateConfig(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_ListProducts0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListProductsRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		out, err := srv.ListProducts(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_UpdateProduct0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateProductRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		out, err := srv.UpdateProduct(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_ListOrders0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListOrdersRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		out, err := srv.ListOrders(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_UpdateOrder0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateOrderRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		out, err := srv.UpdateOrder(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

func _Admin_TriggerSettlement0_HTTP_Handler(srv AdminHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in TriggerSettlementRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		out, err := srv.TriggerSettlement(ctx.Request().Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(200, out)
	}
}

type AdminHTTPClient interface {
	ListUsers(ctx context.Context, req *ListUsersRequest, opts ...http.CallOption) (*ListUsersReply, error)
}

type AdminHTTPClientImpl struct {
	cc *http.Client
}

func NewAdminHTTPClient(client *http.Client) AdminHTTPClient {
	return &AdminHTTPClientImpl{client}
}

func (c *AdminHTTPClientImpl) ListUsers(ctx context.Context, in *ListUsersRequest, opts ...http.CallOption) (*ListUsersReply, error) {
	var out ListUsersReply
	pattern := "/v1/admin/users"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation("/admin.v1.Admin/ListUsers"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	return &out, err
}
