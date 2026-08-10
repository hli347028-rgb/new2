package service

import (
	"context"

	v1 "backend/api/auth/v1"
	"backend/internal/biz"
)

// AuthService implements auth.v1.AuthServer.
type AuthService struct {
	v1.UnimplementedAuthServer

	uc *biz.AuthUsecase
}

// NewAuthService creates AuthService.
func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) GetChallenge(ctx context.Context, req *v1.GetChallengeRequest) (*v1.GetChallengeReply, error) {
	challenge, err := s.uc.GetChallenge(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	return &v1.GetChallengeReply{
		Message:  challenge.Message,
		ExpireAt: challenge.ExpireAt.Unix(),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	user, token, expireAt, isNewUser, err := s.uc.Login(ctx, req.Address, req.Signature, req.InviteCode)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token:          token,
		ExpireAt:       expireAt.Unix(),
		IsNewUser:      isNewUser,
		Address:        user.Address,
		InviterAddress: user.InviterAddress,
		IsAdmin:        s.uc.IsAdminUser(user),
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, req *v1.GetProfileRequest) (*v1.GetProfileReply, error) {
	user, inviteeCount, downline, err := s.uc.GetProfile(ctx, resolveToken(ctx, req.Token))
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DownlineInvitee, 0, len(downline))
	for _, item := range downline {
		items = append(items, &v1.DownlineInvitee{
			Address:    item.Address,
			Generation: item.Generation,
			CreatedAt:  item.CreatedAt.Unix(),
		})
	}
	return &v1.GetProfileReply{
		Address:            user.Address,
		InviterAddress:     user.InviterAddress,
		InviteeCount:       inviteeCount,
		CreatedAt:          user.CreatedAt.Unix(),
		DownlineInvitees:   items,
		TotalDownlineCount: int32(len(items)),
		CommunityLevel:     user.CommunityLevel,
		CommunityStake:     user.CommunityStake,
		TeamStake:          user.TeamStake,
		ShareProfitTotal:   user.ShareProfitTotal,
		EcoRewardTotal:     user.EcoRewardTotal,
		IsAdmin:            s.uc.IsAdminUser(user),
	}, nil
}

func (s *AuthService) ListInvitees(ctx context.Context, req *v1.ListInviteesRequest) (*v1.ListInviteesReply, error) {
	invitees, err := s.uc.ListInvitees(ctx, resolveToken(ctx, req.Token), req.Address)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.InviteeNode, 0, len(invitees))
	for _, item := range invitees {
		items = append(items, &v1.InviteeNode{
			Address:          item.Address,
			TeamStake:        item.TeamStake,
			ReleasedBalance:  item.ReleasedBalance,
			ShareProfitTotal: item.ShareProfitTotal,
			EcoRewardTotal:   item.EcoRewardTotal,
			ExitAmount:       item.ExitAmount,
			CommunityLevel:   item.CommunityLevel,
			DirectCount:      item.DirectCount,
			CreatedAt:        item.CreatedAt.Unix(),
		})
	}
	return &v1.ListInviteesReply{Invitees: items}, nil
}
