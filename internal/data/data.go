package data

import (
	"encoding/json"
	"time"

	"backend/internal/biz"
	"backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewChallengeRepo, NewWalletRepo, NewStakingRepo, NewSettingsRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(dbCfg *conf.DatabaseConfig, logger log.Logger) (*Data, func(), error) {
	db, err := gorm.Open(mysql.Open(dbCfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	if err := db.AutoMigrate(
		&UserPO{}, &OrderPO{}, &RechargePO{}, &TransferPO{}, &WithdrawalPO{},
		&RewardLogPO{}, &MgmtRewardPO{}, &AixPricePO{}, &WinPricePO{}, &SettlementBatchPO{}, &SettingPO{},
		&ExchangeRecordPO{},
	); err != nil {
		return nil, nil, err
	}
	if err := ensureSettlementBatchMultiPerDay(db); err != nil {
		return nil, nil, err
	}
	if err := migrateOverflowReward(db); err != nil {
		return nil, nil, err
	}
	if err := seedDefaults(db); err != nil {
		return nil, nil, err
	}
	if err := ensureZeroAddressAdmin(db); err != nil {
		return nil, nil, err
	}
	if err := refreshAllPerformance(db); err != nil {
		return nil, nil, err
	}
	data := &Data{db: db}
	cleanup := func() {
		sqlDB, err := data.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

// ensureSettlementBatchMultiPerDay removes the historical unique constraint
// on settlement_date so every manual settlement can retain its own batch and
// reward base. Automatic settlement still checks for an existing successful
// date before creating a batch.
func ensureSettlementBatchMultiPerDay(db *gorm.DB) error {
	type indexRow struct {
		IndexName string `gorm:"column:INDEX_NAME"`
	}
	var uniqueIndexes []indexRow
	if err := db.Raw(`
		SELECT INDEX_NAME
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = 'settlement_batches'
		  AND column_name = 'settlement_date'
		  AND NON_UNIQUE = 0
	`).Scan(&uniqueIndexes).Error; err != nil {
		return err
	}
	for _, index := range uniqueIndexes {
		if index.IndexName == "" || index.IndexName == "PRIMARY" {
			continue
		}
		if err := db.Exec("ALTER TABLE settlement_batches DROP INDEX `" + index.IndexName + "`").Error; err != nil {
			return err
		}
	}

	var normalIndexCount int64
	if err := db.Raw(`
		SELECT COUNT(1)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = 'settlement_batches'
		  AND column_name = 'settlement_date'
		  AND NON_UNIQUE = 1
	`).Scan(&normalIndexCount).Error; err != nil {
		return err
	}
	if normalIndexCount == 0 {
		return db.Exec("CREATE INDEX idx_settlement_batches_settlement_date ON settlement_batches (settlement_date)").Error
	}
	return nil
}

// migrateOverflowReward 将历史 pending_mgmt_reward 迁入 overflow_reward，并保持两列同步。
func migrateOverflowReward(db *gorm.DB) error {
	return db.Exec(`
		UPDATE users
		SET overflow_reward = pending_mgmt_reward
		WHERE overflow_reward = 0 AND pending_mgmt_reward > 0
	`).Error
}

// DB exposes the underlying gorm handle for admin legacy queries.
func (d *Data) DB() *gorm.DB {
	return d.db
}

func seedDefaults(db *gorm.DB) error {
	var cnt int64
	if err := db.Model(&SettingPO{}).Where("`key` = ?", conf.SettingsKeySystemConfig).Count(&cnt).Error; err != nil {
		return err
	}
	snap := conf.SystemConfigSnapshot{
		StaticRate:           conf.DefaultStaticRate,
		ExitMultiplier:       conf.DefaultExitMultiplier,
		DirectRate:           conf.DefaultDirectRate,
		MgmtThresholds:       conf.DefaultMgmtThresholds(),
		MgmtRates:            conf.DefaultMgmtRates(),
		AixPriceInitial:      conf.DefaultAixPrice,
		WinPrice:             conf.DefaultWinPrice,
		MgmtCountsTowardExit: true,
		MinSubscribe:         conf.DefaultMinSubscribe,
	}
	if cnt == 0 {
		conf.NormalizeBusinessDefaults(&snap)
		raw, _ := json.Marshal(snap)
		if err := db.Create(&SettingPO{Key: conf.SettingsKeySystemConfig, Value: string(raw)}).Error; err != nil {
			return err
		}
	} else {
		var po SettingPO
		if err := db.Where("`key` = ?", conf.SettingsKeySystemConfig).First(&po).Error; err == nil && po.Value != "" {
			_ = json.Unmarshal([]byte(po.Value), &snap)
			conf.NormalizeBusinessDefaults(&snap)
		}
	}
	today := time.Now().Format("2006-01-02")
	aixSeed := decimal.NewFromFloat(snap.AixPriceInitial)
	if !aixSeed.IsPositive() {
		aixSeed = decimal.NewFromFloat(conf.DefaultAixPrice)
	}
	var priceCnt int64
	if err := db.Model(&AixPricePO{}).Where("effective_date = ?", today).Count(&priceCnt).Error; err != nil {
		return err
	}
	if priceCnt == 0 {
		if err := db.Create(&AixPricePO{
			Price:         aixSeed,
			EffectiveDate: today,
			Remark:        "initial",
		}).Error; err != nil {
			return err
		}
	}
	var winCnt int64
	if err := db.Model(&WinPricePO{}).Where("id = ?", WinPriceRowID).Count(&winCnt).Error; err != nil {
		return err
	}
	if winCnt == 0 {
		winSeed := decimal.NewFromFloat(snap.WinPrice)
		if !winSeed.IsPositive() {
			winSeed = decimal.NewFromFloat(conf.DefaultWinPrice)
		}
		return db.Create(&WinPricePO{
			ID:     WinPriceRowID,
			Price:  winSeed,
			Source: "initial",
		}).Error
	}
	return nil
}

func ensureZeroAddressAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&UserPO{}).Where("address = ?", biz.ZeroAddress).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return db.Create(&UserPO{
			Address:    biz.ZeroAddress,
			InviteCode: biz.ZeroAddress,
			Role:       biz.RoleAdmin,
			Status:     1,
		}).Error
	}
	return db.Model(&UserPO{}).
		Where("address = ?", biz.ZeroAddress).
		Update("role", biz.RoleAdmin).Error
}
