package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"backend/internal/biz"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) FindByAddress(ctx context.Context, address string) (*biz.User, error) {
	var po UserPO
	err := r.data.db.WithContext(ctx).Where("address = ?", address).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toBiz(ctx, &po), nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	var po UserPO
	err := r.data.db.WithContext(ctx).First(&po, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toBiz(ctx, &po), nil
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	invite := user.InviteCode
	if invite == "" {
		invite = user.Address
	}
	po := &UserPO{
		Address:    user.Address,
		InviterID:  user.InviterID,
		InviteCode: invite,
		Role:       biz.RoleUser,
		Status:     1,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return r.toBiz(ctx, po), nil
}

func (r *userRepo) CountInvitees(ctx context.Context, userID int64) (int32, error) {
	var count int64
	err := r.data.db.WithContext(ctx).Model(&UserPO{}).Where("inviter_id = ?", userID).Count(&count).Error
	return int32(count), err
}

func (r *userRepo) ListDownlineInvitees(ctx context.Context, userID int64, maxDepth int) ([]*biz.DownlineInvitee, error) {
	if maxDepth <= 0 {
		maxDepth = 1000
	}
	var result []*biz.DownlineInvitee
	currentLevelIDs := []int64{userID}
	for generation := 1; generation <= maxDepth; generation++ {
		if len(currentLevelIDs) == 0 {
			break
		}
		var pos []UserPO
		if err := r.data.db.WithContext(ctx).
			Where("inviter_id IN ?", currentLevelIDs).
			Order("created_time ASC").
			Find(&pos).Error; err != nil {
			return nil, err
		}
		if len(pos) == 0 {
			break
		}
		next := make([]int64, 0, len(pos))
		for _, po := range pos {
			result = append(result, &biz.DownlineInvitee{
				Address:    po.Address,
				Generation: int32(generation),
				CreatedAt:  po.CreatedTime,
			})
			next = append(next, po.ID)
		}
		currentLevelIDs = next
	}
	return result, nil
}

func (r *userRepo) ListUsersUnder(ctx context.Context, rootID int64) ([]*biz.User, error) {
	pos, err := r.listUsersUnder(ctx, rootID)
	if err != nil {
		return nil, err
	}
	out := make([]*biz.User, 0, len(pos))
	for i := range pos {
		out = append(out, r.toBiz(ctx, &pos[i]))
	}
	return out, nil
}

func (r *userRepo) listUsersUnder(ctx context.Context, rootID int64) ([]UserPO, error) {
	var all []UserPO
	current := []int64{rootID}
	for depth := 0; depth < 1000 && len(current) > 0; depth++ {
		var pos []UserPO
		if err := r.data.db.WithContext(ctx).
			Where("inviter_id IN ?", current).
			Order("id asc").
			Find(&pos).Error; err != nil {
			return nil, err
		}
		if len(pos) == 0 {
			break
		}
		all = append(all, pos...)
		next := make([]int64, 0, len(pos))
		for _, po := range pos {
			next = append(next, po.ID)
		}
		current = next
	}
	return all, nil
}

func (r *userRepo) ListAllUsers(ctx context.Context) ([]*biz.User, error) {
	var list []UserPO
	if err := r.data.db.WithContext(ctx).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.User, 0, len(list))
	for i := range list {
		out = append(out, r.toBiz(ctx, &list[i]))
	}
	return out, nil
}

func (r *userRepo) ListDirectInvitees(ctx context.Context, userID int64) ([]*biz.User, error) {
	var list []UserPO
	if err := r.data.db.WithContext(ctx).Where("inviter_id = ?", userID).Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.User, 0, len(list))
	for i := range list {
		out = append(out, r.toBiz(ctx, &list[i]))
	}
	return out, nil
}

func (r *userRepo) SumPrincipalByUserIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(userIDs))
	for _, id := range userIDs {
		out[id] = "0"
	}
	if len(userIDs) == 0 {
		return out, nil
	}
	type row struct {
		UserID int64
		Total  decimal.Decimal
	}
	var rows []row
	err := r.data.db.WithContext(ctx).Raw(`
		SELECT user_id, COALESCE(SUM(principal), 0) AS total
		FROM orders
		WHERE user_id IN ? AND status = ?
		GROUP BY user_id
	`, userIDs, biz.OrderStatusActive).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		out[item.UserID] = item.Total.String()
	}
	return out, nil
}

func (r *userRepo) SumExitAmountByUserIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	return r.SumPrincipalByUserIDs(ctx, userIDs)
}

func (r *userRepo) UpdateMgmtStats(ctx context.Context, userID int64, level int32, smallArea, teamPerf string) error {
	sa, err := decimal.NewFromString(smallArea)
	if err != nil {
		return err
	}
	tp, err := decimal.NewFromString(teamPerf)
	if err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"mgmt_level":      level,
		"small_area_perf": sa,
		"team_perf":       tp,
	}).Error
}

func (r *userRepo) AdminUpdateUser(ctx context.Context, update *biz.AdminUserUpdate) error {
	updates := map[string]interface{}{}
	if update.UsdtRecharge != "" {
		v, err := decimal.NewFromString(update.UsdtRecharge)
		if err != nil {
			return err
		}
		updates["usdt_recharge"] = v
	}
	if update.UsdtReward != "" {
		v, err := decimal.NewFromString(update.UsdtReward)
		if err != nil {
			return err
		}
		updates["usdt_reward"] = v
	}
	if update.AixBalance != "" {
		v, err := decimal.NewFromString(update.AixBalance)
		if err != nil {
			return err
		}
		updates["aix_balance"] = v
	}
	if update.StaticUsdtTotal != "" {
		v, err := decimal.NewFromString(update.StaticUsdtTotal)
		if err != nil {
			return err
		}
		updates["static_usdt_total"] = v
	}
	if update.Balance != "" {
		v, err := decimal.NewFromString(update.Balance)
		if err != nil {
			return err
		}
		updates["usdt_recharge"] = v
	}
	if update.ReleasedBalance != "" {
		v, err := decimal.NewFromString(update.ReleasedBalance)
		if err != nil {
			return err
		}
		updates["usdt_reward"] = v
	}
	if update.Role != "" {
		updates["role"] = update.Role
	}
	if update.SetCommunityLevel {
		// parse W0-W10 or bare 0-10
		level := int32(0)
		lv := strings.TrimSpace(update.CommunityLevel)
		lv = strings.TrimPrefix(strings.ToUpper(lv), "W")
		lv = strings.TrimPrefix(lv, "V")
		if n, err := strconv.Atoi(lv); err == nil && n >= 0 {
			if n > 10 {
				n = 10
			}
			level = int32(n)
		}
		updates["mgmt_level"] = level
	}
	if update.CommunityStake != "" {
		v, err := decimal.NewFromString(update.CommunityStake)
		if err != nil {
			return err
		}
		updates["small_area_perf"] = v
	}
	if update.TeamStake != "" {
		v, err := decimal.NewFromString(update.TeamStake)
		if err != nil {
			return err
		}
		updates["team_perf"] = v
	}
	if update.InviterID != nil {
		updates["inviter_id"] = update.InviterID
	}
	if len(updates) == 0 {
		return nil
	}
	return r.data.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", update.UserID).Updates(updates).Error
}

func (r *userRepo) SetRole(ctx context.Context, userID int64, role string) error {
	return r.data.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", userID).Update("role", role).Error
}

func (r *userRepo) GetBalances(ctx context.Context, userID int64) (string, string, string, error) {
	var po UserPO
	if err := r.data.db.WithContext(ctx).Select("usdt_recharge", "usdt_reward", "aix_balance").First(&po, userID).Error; err != nil {
		return "", "", "", err
	}
	return po.UsdtRecharge.String(), po.UsdtReward.String(), po.AixBalance.String(), nil
}

func (r *userRepo) AddUsdtRecharge(ctx context.Context, userID int64, amount string) (string, error) {
	amountDec, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	var newBal decimal.Decimal
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, userID).Error; err != nil {
			return err
		}
		po.UsdtRecharge = po.UsdtRecharge.Add(amountDec)
		newBal = po.UsdtRecharge
		return tx.Model(&po).Update("usdt_recharge", po.UsdtRecharge).Error
	})
	return newBal.String(), err
}

func (r *userRepo) IsUplineOrDownline(ctx context.Context, a, b int64) (bool, error) {
	if a == b {
		return false, nil
	}
	var users []UserPO
	if err := r.data.db.WithContext(ctx).Select("id", "inviter_id").Find(&users).Error; err != nil {
		return false, err
	}
	parents := make(map[int64]int64, len(users))
	for _, user := range users {
		if user.InviterID != nil {
			parents[user.ID] = *user.InviterID
		}
	}
	return biz.IsLinealRelation(a, b, parents), nil
}

func (r *userRepo) SetWithdrawReset(ctx context.Context, userID int64, reset bool) error {
	return nil
}
func (r *userRepo) ClearWithdrawReset(ctx context.Context, userID int64) error {
	return nil
}
func (r *userRepo) IsWithdrawReset(ctx context.Context, userID int64) (bool, error) {
	return false, nil
}
func (r *userRepo) UpdateCommunityStats(ctx context.Context, userID int64, level string, communityStake, teamStake string) error {
	lv := int32(0)
	if len(level) >= 2 && (level[0] == 'W' || level[0] == 'w') {
		n := 0
		for i := 1; i < len(level); i++ {
			if level[i] >= '0' && level[i] <= '9' {
				n = n*10 + int(level[i]-'0')
			}
		}
		lv = int32(n)
	}
	return r.UpdateMgmtStats(ctx, userID, lv, communityStake, teamStake)
}
func (r *userRepo) GetBalance(ctx context.Context, userID int64) (string, error) {
	recharge, _, _, err := r.GetBalances(ctx, userID)
	return recharge, err
}
func (r *userRepo) GetReleasedBalance(ctx context.Context, userID int64) (string, error) {
	_, reward, _, err := r.GetBalances(ctx, userID)
	return reward, err
}
func (r *userRepo) AddBalance(ctx context.Context, userID int64, amount string) (string, error) {
	return r.AddUsdtRecharge(ctx, userID, amount)
}
func (r *userRepo) AddReleasedBalance(ctx context.Context, userID int64, amount string) (string, error) {
	amountDec, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	var newBal decimal.Decimal
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, userID).Error; err != nil {
			return err
		}
		po.UsdtReward = po.UsdtReward.Add(amountDec)
		newBal = po.UsdtReward
		return tx.Model(&po).Update("usdt_reward", po.UsdtReward).Error
	})
	return newBal.String(), err
}
func (r *userRepo) ClaimReleasedToAccount(ctx context.Context, userID int64, amount string) (string, string, error) {
	return "", "", fmt.Errorf("not supported in AIX")
}
func (r *userRepo) DeductBalance(ctx context.Context, userID int64, amount string) (string, error) {
	amountDec, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	var newBal decimal.Decimal
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po UserPO
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, userID).Error; err != nil {
			return err
		}
		if po.UsdtRecharge.LessThan(amountDec) {
			return fmt.Errorf("insufficient balance")
		}
		po.UsdtRecharge = po.UsdtRecharge.Sub(amountDec)
		newBal = po.UsdtRecharge
		return tx.Model(&po).Update("usdt_recharge", po.UsdtRecharge).Error
	})
	return newBal.String(), err
}

func (r *userRepo) toBiz(ctx context.Context, po *UserPO) *biz.User {
	user := &biz.User{
		ID:              po.ID,
		Address:         po.Address,
		InviteCode:      po.InviteCode,
		UsdtRecharge:    po.UsdtRecharge.String(),
		UsdtReward:      po.UsdtReward.String(),
		AixBalance:      po.AixBalance.String(),
		StaticUsdtTotal: po.StaticUsdtTotal.String(),
		MgmtLevel:       po.MgmtLevel,
		LargeAreaPerf:   po.LargeAreaPerf.String(),
		SmallAreaPerf:   po.SmallAreaPerf.String(),
		TeamPerf:        po.TeamPerf.String(),
		Status:          po.Status,
		InviterID:       po.InviterID,
		Role:            po.Role,
		CreatedTime:     po.CreatedTime,
		UpdatedTime:     po.UpdatedTime,
	}
	if po.InviterID != nil {
		var addr string
		if err := r.data.db.WithContext(ctx).Model(&UserPO{}).
			Select("address").Where("id = ?", *po.InviterID).Scan(&addr).Error; err == nil {
			user.InviterAddress = addr
		}
	}
	user.SyncCompatFields()
	return user
}
