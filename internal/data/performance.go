package data

import (
	"context"
	"fmt"

	"backend/internal/biz"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// refreshPerformance recalculates cached performance from active order
// principal. A direct referral and all of its descendants form one branch.
func refreshPerformance(db *gorm.DB, sourceUserID int64) error {
	type userNode struct {
		ID              int64
		InviterID       *int64
		MgmtLevelLocked bool
	}
	type principalRow struct {
		UserID int64
		Total  decimal.Decimal
	}

	var users []userNode
	if err := db.Model(&UserPO{}).Select("id", "inviter_id", "mgmt_level_locked").Find(&users).Error; err != nil {
		return err
	}
	var principals []principalRow
	if err := db.Model(&OrderPO{}).
		Select("user_id, COALESCE(SUM(principal), 0) AS total").
		Where("status = ?", biz.OrderStatusActive).
		Group("user_id").Scan(&principals).Error; err != nil {
		return err
	}

	children := make(map[int64][]int64, len(users))
	parents := make(map[int64]int64, len(users))
	userByID := make(map[int64]userNode, len(users))
	stake := make(map[int64]decimal.Decimal, len(users))
	for _, user := range users {
		userByID[user.ID] = user
		stake[user.ID] = decimal.Zero
		if user.InviterID != nil {
			children[*user.InviterID] = append(children[*user.InviterID], user.ID)
			parents[user.ID] = *user.InviterID
		}
	}
	for _, row := range principals {
		stake[row.UserID] = row.Total
	}

	memo := make(map[int64]decimal.Decimal, len(users))
	visiting := make(map[int64]bool, len(users))
	var subtree func(int64) (decimal.Decimal, error)
	subtree = func(id int64) (decimal.Decimal, error) {
		if value, ok := memo[id]; ok {
			return value, nil
		}
		if visiting[id] {
			return decimal.Zero, fmt.Errorf("invite relationship contains a cycle at user %d", id)
		}
		visiting[id] = true
		total := stake[id]
		for _, childID := range children[id] {
			childTotal, err := subtree(childID)
			if err != nil {
				return decimal.Zero, err
			}
			total = total.Add(childTotal)
		}
		visiting[id] = false
		memo[id] = total
		return total, nil
	}

	targets := users
	if sourceUserID > 0 {
		targets = make([]userNode, 0)
		seen := map[int64]bool{sourceUserID: true}
		currentID := sourceUserID
		for {
			parentID, ok := parents[currentID]
			if !ok {
				break
			}
			if seen[parentID] {
				return fmt.Errorf("invite relationship contains a cycle at user %d", parentID)
			}
			seen[parentID] = true
			targets = append(targets, userByID[parentID])
			currentID = parentID
		}
	}

	for _, user := range targets {
		branches := make([]decimal.Decimal, 0, len(children[user.ID]))
		for _, childID := range children[user.ID] {
			value, err := subtree(childID)
			if err != nil {
				return err
			}
			branches = append(branches, value)
		}
		large, small, team := biz.CalcAreaPerformance(branches)
		level := biz.MgmtLevelByPerf(small)
		updates := map[string]interface{}{
			"large_area_perf": large,
			"small_area_perf": small,
			"team_perf":       team,
		}
		if !user.MgmtLevelLocked {
			updates["mgmt_level"] = level
		}
		if err := db.Model(&UserPO{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func refreshAllPerformance(db *gorm.DB) error {
	return refreshPerformance(db, 0)
}

func refreshAncestorPerformance(db *gorm.DB, sourceUserID int64) error {
	return refreshPerformance(db, sourceUserID)
}

func (r *userRepo) RefreshPerformance(ctx context.Context) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return refreshAllPerformance(tx)
	})
}
