package biz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	v1 "backend/api/auth/v1"
	"backend/internal/conf"
	"backend/internal/pkg/eth"
	"backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type AuthUsecase struct {
	userRepo      UserRepo
	challengeRepo ChallengeRepo
	auth          *conf.AuthConfig
	log           *log.Helper
}

func NewAuthUsecase(userRepo UserRepo, challengeRepo ChallengeRepo, auth *conf.AuthConfig, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		userRepo:      userRepo,
		challengeRepo: challengeRepo,
		auth:          auth,
		log:           log.NewHelper(logger),
	}
}

func (uc *AuthUsecase) challengeTTL() time.Duration {
	return uc.auth.ChallengeDuration()
}

func (uc *AuthUsecase) jwtSecret() string {
	return uc.auth.GetJwtSecret()
}

func (uc *AuthUsecase) isBootstrapAddress(address string) bool {
	for _, item := range uc.auth.GetBootstrapAddresses() {
		normalized, err := eth.NormalizeAddress(item)
		if err == nil && normalized == address {
			return true
		}
	}
	return false
}

func (uc *AuthUsecase) GetChallenge(ctx context.Context, address string) (*Challenge, error) {
	normalized, err := eth.NormalizeAddress(address)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ADDRESS.String(), "无效的钱包地址")
	}
	nonce, err := randomNonce(16)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expireAt := now.Add(uc.challengeTTL())
	message := fmt.Sprintf(
		"Welcome to AIX Web3 login\nAddress: %s\nNonce: %s\nTimestamp: %d",
		normalized, nonce, now.Unix(),
	)
	challenge := &Challenge{Address: normalized, Message: message, ExpireAt: expireAt}
	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, address, signature, inviteCode string) (*User, string, time.Time, bool, error) {
	normalized, err := eth.NormalizeAddress(address)
	if err != nil {
		return nil, "", time.Time{}, false, errors.BadRequest(v1.ErrorReason_INVALID_ADDRESS.String(), "无效的钱包地址")
	}
	isZeroAdminLogin := normalized == ZeroAddress
	if !isZeroAdminLogin {
		challenge, err := uc.challengeRepo.Get(ctx, normalized)
		if err != nil {
			return nil, "", time.Time{}, false, errors.BadRequest(v1.ErrorReason_CHALLENGE_NOT_FOUND.String(), "请先获取签名挑战")
		}
		if time.Now().After(challenge.ExpireAt) {
			_ = uc.challengeRepo.Delete(ctx, normalized)
			return nil, "", time.Time{}, false, errors.BadRequest(v1.ErrorReason_CHALLENGE_EXPIRED.String(), "签名挑战已过期，请重新获取")
		}
		if err := eth.VerifyPersonalSign(challenge.Message, signature, normalized); err != nil {
			return nil, "", time.Time{}, false, errors.Unauthorized(v1.ErrorReason_INVALID_SIGNATURE.String(), "签名校验失败")
		}
		_ = uc.challengeRepo.Delete(ctx, normalized)
	}

	existing, err := uc.userRepo.FindByAddress(ctx, normalized)
	if err != nil {
		return nil, "", time.Time{}, false, err
	}
	isNewUser := existing == nil
	var user *User
	if isNewUser {
		user, err = uc.registerNewUser(ctx, normalized, inviteCode)
		if err != nil {
			return nil, "", time.Time{}, false, err
		}
	} else {
		user = existing
	}

	jwtToken, expireAt, err := token.Generate(normalized, uc.jwtSecret(), time.Now())
	if err != nil {
		return nil, "", time.Time{}, false, err
	}
	if uc.shouldBeAdmin(normalized, user) {
		_ = uc.userRepo.SetRole(ctx, user.ID, RoleAdmin)
		user.Role = RoleAdmin
	}
	return user, jwtToken, expireAt, isNewUser, nil
}

func (uc *AuthUsecase) shouldBeAdmin(address string, user *User) bool {
	if user != nil && (user.Role == RoleAdmin || user.Address == ZeroAddress) {
		return true
	}
	if address == ZeroAddress {
		return true
	}
	for _, item := range uc.auth.GetAdminAddresses() {
		if strings.EqualFold(item, address) {
			return true
		}
	}
	return false
}

func (uc *AuthUsecase) IsAdminUser(user *User) bool {
	return IsAdmin(user, uc.auth)
}

func (uc *AuthUsecase) registerNewUser(ctx context.Context, address, inviteCode string) (*User, error) {
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" && !uc.isBootstrapAddress(address) && address != ZeroAddress {
		return nil, errors.BadRequest(v1.ErrorReason_INVITE_CODE_REQUIRED.String(), "首次登录需要邀请码（已登录用户的钱包地址）")
	}
	var inviter *User
	if inviteCode != "" {
		normalizedInvite, err := eth.NormalizeAddress(inviteCode)
		if err != nil {
			return nil, errors.BadRequest(v1.ErrorReason_INVITE_CODE_INVALID.String(), "邀请码格式无效")
		}
		inviter, err = uc.userRepo.FindByAddress(ctx, normalizedInvite)
		if err != nil {
			return nil, err
		}
		if inviter == nil {
			return nil, errors.BadRequest(v1.ErrorReason_INVITE_CODE_INVALID.String(), "邀请码无效，邀请人尚未登录注册")
		}
	}
	user := &User{Address: address, InviteCode: address}
	if inviter != nil {
		user.InviterID = &inviter.ID
		user.InviterAddress = inviter.Address
	}
	return uc.userRepo.Create(ctx, user)
}

func (uc *AuthUsecase) GetProfile(ctx context.Context, tokenString string) (*User, int32, []*DownlineInvitee, error) {
	address, err := token.Parse(tokenString, uc.jwtSecret())
	if err != nil {
		return nil, 0, nil, errors.Unauthorized(v1.ErrorReason_AUTH_UNSPECIFIED.String(), "token 无效或已过期")
	}
	user, err := uc.userRepo.FindByAddress(ctx, address)
	if err != nil {
		return nil, 0, nil, err
	}
	if user == nil {
		return nil, 0, nil, errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "用户不存在")
	}
	count, err := uc.userRepo.CountInvitees(ctx, user.ID)
	if err != nil {
		return nil, 0, nil, err
	}
	downline, err := uc.userRepo.ListDownlineInvitees(ctx, user.ID, MaxDownlineGenerations)
	if err != nil {
		return nil, 0, nil, err
	}
	return user, count, downline, nil
}

func (uc *AuthUsecase) ListInvitees(ctx context.Context, tokenString, address string) ([]*DirectInvitee, error) {
	selfAddress, err := token.Parse(tokenString, uc.jwtSecret())
	if err != nil {
		return nil, errors.Unauthorized(v1.ErrorReason_AUTH_UNSPECIFIED.String(), "token 无效或已过期")
	}
	targetAddress := strings.TrimSpace(address)
	if targetAddress == "" {
		targetAddress = selfAddress
	}
	normalized, err := eth.NormalizeAddress(targetAddress)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ADDRESS.String(), "无效的钱包地址")
	}
	user, err := uc.userRepo.FindByAddress(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "用户不存在")
	}
	invitees, err := uc.userRepo.ListDirectInvitees(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	stakeMap, childrenMap, err := uc.buildPerfTree(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	memo := make(map[int64]decimal.Decimal)
	result := make([]*DirectInvitee, 0, len(invitees))
	for _, item := range invitees {
		directCount, err := uc.userRepo.CountInvitees(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		lineExit := CalcSubtreeStake(item.ID, stakeMap, childrenMap, memo)
		result = append(result, &DirectInvitee{
			Address:          item.Address,
			TeamStake:        item.TeamPerf,
			ExitAmount:       lineExit.String(),
			CommunityLevel:   item.CommunityLevel,
			ReleasedBalance:  item.UsdtReward,
			ShareProfitTotal: "0",
			EcoRewardTotal:   "0",
			DirectCount:      directCount,
			CreatedAt:        item.CreatedTime,
		})
	}
	return result, nil
}

func (uc *AuthUsecase) buildPerfTree(ctx context.Context, rootID int64) (map[int64]decimal.Decimal, map[int64][]int64, error) {
	under, err := uc.userRepo.ListUsersUnder(ctx, rootID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]int64, 0, len(under)+1)
	ids = append(ids, rootID)
	childrenMap := make(map[int64][]int64)
	for _, u := range under {
		ids = append(ids, u.ID)
		if u.InviterID != nil {
			childrenMap[*u.InviterID] = append(childrenMap[*u.InviterID], u.ID)
		}
	}
	perfMap, err := uc.userRepo.SumPrincipalByUserIDs(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	stakeMap := make(map[int64]decimal.Decimal, len(ids))
	for _, id := range ids {
		v, _ := decimal.NewFromString(perfMap[id])
		stakeMap[id] = v
	}
	return stakeMap, childrenMap, nil
}

func randomNonce(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
