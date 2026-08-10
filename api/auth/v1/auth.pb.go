package v1

type ErrorReason int32

const (
	ErrorReason_AUTH_UNSPECIFIED    ErrorReason = 0
	ErrorReason_INVALID_ADDRESS     ErrorReason = 1
	ErrorReason_INVALID_SIGNATURE   ErrorReason = 2
	ErrorReason_CHALLENGE_NOT_FOUND ErrorReason = 3
	ErrorReason_CHALLENGE_EXPIRED   ErrorReason = 4
	ErrorReason_INVITE_CODE_REQUIRED ErrorReason = 5
	ErrorReason_INVITE_CODE_INVALID ErrorReason = 6
	ErrorReason_USER_NOT_FOUND      ErrorReason = 7
)

func (e ErrorReason) String() string {
	switch e {
	case ErrorReason_INVALID_ADDRESS:
		return "INVALID_ADDRESS"
	case ErrorReason_INVALID_SIGNATURE:
		return "INVALID_SIGNATURE"
	case ErrorReason_CHALLENGE_NOT_FOUND:
		return "CHALLENGE_NOT_FOUND"
	case ErrorReason_CHALLENGE_EXPIRED:
		return "CHALLENGE_EXPIRED"
	case ErrorReason_INVITE_CODE_REQUIRED:
		return "INVITE_CODE_REQUIRED"
	case ErrorReason_INVITE_CODE_INVALID:
		return "INVITE_CODE_INVALID"
	case ErrorReason_USER_NOT_FOUND:
		return "USER_NOT_FOUND"
	default:
		return "AUTH_UNSPECIFIED"
	}
}

type GetChallengeRequest struct {
	Address string `json:"address"`
}

type GetChallengeReply struct {
	Message  string `json:"message"`
	ExpireAt int64  `json:"expire_at"`
}

type LoginRequest struct {
	Address    string `json:"address"`
	Signature  string `json:"signature"`
	InviteCode string `json:"invite_code"`
}

type LoginReply struct {
	Token          string `json:"token"`
	ExpireAt       int64  `json:"expire_at"`
	IsNewUser      bool   `json:"is_new_user"`
	Address        string `json:"address"`
	InviterAddress string `json:"inviter_address"`
	IsAdmin        bool   `json:"is_admin"`
}

type GetProfileRequest struct {
	Token string `json:"token"`
}

type GetProfileReply struct {
	Address            string             `json:"address"`
	InviterAddress     string             `json:"inviter_address"`
	InviteeCount       int32              `json:"invitee_count"`
	CreatedAt          int64              `json:"created_at"`
	DownlineInvitees   []*DownlineInvitee `json:"downline_invitees"`
	TotalDownlineCount int32              `json:"total_downline_count"`
	CommunityLevel     string             `json:"community_level"`
	CommunityStake     string             `json:"community_stake"`
	TeamStake          string             `json:"team_stake"`
	ShareProfitTotal   string             `json:"share_profit_total"`
	EcoRewardTotal     string             `json:"eco_reward_total"`
	IsAdmin            bool               `json:"is_admin"`
}

type DownlineInvitee struct {
	Address    string `json:"address"`
	Generation int32  `json:"generation"`
	CreatedAt  int64  `json:"created_at"`
}

type ListInviteesRequest struct {
	Token   string `json:"token"`
	Address string `json:"address"`
}

type InviteeNode struct {
	Address          string `json:"address"`
	TeamStake        string `json:"team_stake"`
	ReleasedBalance  string `json:"released_balance"`
	ShareProfitTotal string `json:"share_profit_total"`
	EcoRewardTotal   string `json:"eco_reward_total"`
	ExitAmount       string `json:"exit_amount"`
	CommunityLevel   string `json:"community_level"`
	DirectCount      int32  `json:"direct_count"`
	CreatedAt        int64  `json:"created_at"`
}

type ListInviteesReply struct {
	Invitees []*InviteeNode `json:"invitees"`
}
