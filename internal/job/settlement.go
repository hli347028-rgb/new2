package job

import (
	"context"
	"time"

	"backend/internal/biz"
	"backend/internal/pkg/token"

	"github.com/go-kratos/kratos/v2/log"
)

// SettlementJob 每日中国时间 0 点结算
type SettlementJob struct {
	uc     *biz.SettlementUsecase
	log    *log.Helper
	stopCh chan struct{}
}

func NewSettlementJob(uc *biz.SettlementUsecase, logger log.Logger) *SettlementJob {
	return &SettlementJob{
		uc:     uc,
		log:    log.NewHelper(logger),
		stopCh: make(chan struct{}),
	}
}

func (j *SettlementJob) Start() {
	go j.run()
	j.log.Info("settlement job started, runs daily at China midnight")
}

func (j *SettlementJob) Stop() {
	close(j.stopCh)
}

func (j *SettlementJob) run() {
	j.runOnce()

	for {
		delay := durationUntilNextChinaMidnight(time.Now())
		j.log.Infof("next settlement in %s", delay.Round(time.Second))
		select {
		case <-time.After(delay):
			j.runOnce()
		case <-j.stopCh:
			return
		}
	}
}

func (j *SettlementJob) runOnce() {
	ctx := context.Background()
	settlementDate := biz.TodaySettlementDate(time.Now())
	j.log.Infof("running daily settlement for %s", settlementDate)
	if err := j.uc.RunDailySettlement(ctx, settlementDate); err != nil {
		j.log.Errorf("settlement %s failed: %v", settlementDate, err)
		return
	}
	j.log.Infof("settlement %s completed", settlementDate)

	// 补发历史漏掉的社区基础奖（有社区等级但 eco_rewards 为空的日期）
	if err := j.uc.BackfillMissingEcoRewards(ctx); err != nil {
		j.log.Errorf("eco backfill failed: %v", err)
	}
}

func durationUntilNextChinaMidnight(now time.Time) time.Duration {
	loc := token.ChinaLocation()
	now = now.In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	return next.Sub(now)
}
