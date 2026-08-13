package biz

import (
	"context"
	"time"
)

// User represents an AIX wallet user.
type User struct {
	ID              int64
	Address         string
	InviteCode      string
	UsdtRecharge      string
	UsdtReward        string
	AixBalance        string // AIX 代币数
	WinBalance        string // WIN 代币数
	PendingMgmtReward string // 兼容旧字段 = OverflowReward
	OverflowReward    string // 溢出奖励（USDT）
	StaticUsdtTotal   string // 静态总收益（USDT）
	MgmtLevel       int32
	LargeAreaPerf   string
	SmallAreaPerf   string
	TeamPerf        string
	Status          int32
	InviterID       *int64
	InviterAddress  string
	Role            string
	CreatedTime     time.Time
	UpdatedTime     time.Time

	// Compatibility aliases for legacy admin/auth mapping
	Balance              string // = UsdtRecharge
	ReleasedBalance      string // = UsdtReward
	CommunityLevel       string // = W{mgmt_level}
	CommunityStake       string // = SmallAreaPerf
	TeamStake            string // = TeamPerf
	ShareProfitTotal     string
	EcoRewardTotal       string
	CommunityLevelLocked bool
	CreatedAt            time.Time
}

// SyncCompatFields fills legacy alias fields from AIX balances.
func (u *User) SyncCompatFields() {
	if u == nil {
		return
	}
	u.Balance = u.UsdtRecharge
	u.ReleasedBalance = u.UsdtReward
	u.CommunityStake = u.SmallAreaPerf
	u.TeamStake = u.TeamPerf
	u.CreatedAt = u.CreatedTime
	if u.MgmtLevel > 0 {
		u.CommunityLevel = "W" + itoa(int(u.MgmtLevel))
	} else {
		u.CommunityLevel = "W0"
	}
	if u.ShareProfitTotal == "" {
		u.ShareProfitTotal = "0"
	}
	if u.EcoRewardTotal == "" {
		u.EcoRewardTotal = "0"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// DownlineInvitee represents a user in the inviter's downline tree.
type DownlineInvitee struct {
	Address    string
	Generation int32
	CreatedAt  time.Time
}

// DirectInvitee 直推下级节点
type DirectInvitee struct {
	Address          string
	TeamStake        string
	ExitAmount       string
	CommunityLevel   string
	ReleasedBalance  string
	ShareProfitTotal string
	EcoRewardTotal   string
	DirectCount      int32
	CreatedAt        time.Time
}

const MaxDownlineGenerations = 10

// Challenge represents a login challenge message.
type Challenge struct {
	Address  string
	Message  string
	ExpireAt time.Time
}

// ChallengeRepo stores temporary login challenges.
type ChallengeRepo interface {
	Save(ctx context.Context, challenge *Challenge) error
	Get(ctx context.Context, address string) (*Challenge, error)
	Delete(ctx context.Context, address string) error
}

// UserRepo handles user persistence.
type UserRepo interface {
	FindByAddress(ctx context.Context, address string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	CountInvitees(ctx context.Context, userID int64) (int32, error)
	ListDownlineInvitees(ctx context.Context, userID int64, maxDepth int) ([]*DownlineInvitee, error)
	ListAllUsers(ctx context.Context) ([]*User, error)
	ListDirectInvitees(ctx context.Context, userID int64) ([]*User, error)
	ListUsersUnder(ctx context.Context, rootID int64) ([]*User, error)
	SumPrincipalByUserIDs(ctx context.Context, userIDs []int64) (map[int64]string, error)
	// SumExitAmountByUserIDs 兼容旧接口：活跃订单本金合计（业绩）
	SumExitAmountByUserIDs(ctx context.Context, userIDs []int64) (map[int64]string, error)
	UpdateMgmtStats(ctx context.Context, userID int64, level int32, smallArea, teamPerf string) error
	RefreshPerformance(ctx context.Context) error
	AdminUpdateUser(ctx context.Context, update *AdminUserUpdate) error
	SetRole(ctx context.Context, userID int64, role string) error
	GetBalances(ctx context.Context, userID int64) (recharge, reward, aix string, err error)
	AddUsdtRecharge(ctx context.Context, userID int64, amount string) (string, error)
	IsUplineOrDownline(ctx context.Context, a, b int64) (bool, error)

	// Legacy stubs kept for compile of unused paths
	SetWithdrawReset(ctx context.Context, userID int64, reset bool) error
	ClearWithdrawReset(ctx context.Context, userID int64) error
	IsWithdrawReset(ctx context.Context, userID int64) (bool, error)
	UpdateCommunityStats(ctx context.Context, userID int64, level string, communityStake, teamStake string) error
	GetBalance(ctx context.Context, userID int64) (string, error)
	GetReleasedBalance(ctx context.Context, userID int64) (string, error)
	AddBalance(ctx context.Context, userID int64, amount string) (string, error)
	AddReleasedBalance(ctx context.Context, userID int64, amount string) (string, error)
	ClaimReleasedToAccount(ctx context.Context, userID int64, amount string) (string, string, error)
	DeductBalance(ctx context.Context, userID int64, amount string) (string, error)
}
