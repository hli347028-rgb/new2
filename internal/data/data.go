package data

import (
	"backend/internal/biz"
	"backend/internal/conf"
	"encoding/json"
	"time"

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
		&RewardLogPO{}, &MgmtRewardPO{}, &AixPricePO{}, &SettlementBatchPO{}, &SettingPO{},
		&ExchangeRecordPO{},
	); err != nil {
		return nil, nil, err
	}
	if err := ensureSettlementBatchMultiPerDay(db); err != nil {
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

// DB exposes the underlying gorm handle for admin legacy queries.
func (d *Data) DB() *gorm.DB {
	return d.db
}

func seedDefaults(db *gorm.DB) error {
	var cnt int64
	if err := db.Model(&SettingPO{}).Where("`key` = ?", conf.SettingsKeySystemConfig).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		snap := conf.SystemConfigSnapshot{
			StaticRate:           conf.DefaultStaticRate,
			ExitMultiplier:       conf.DefaultExitMultiplier,
			DirectRate:           conf.DefaultDirectRate,
			MgmtThresholds:       conf.DefaultMgmtThresholds(),
			MgmtRates:            conf.DefaultMgmtRates(),
			AixPriceInitial:      conf.DefaultAixPrice,
			MgmtCountsTowardExit: true,
			MinSubscribe:         conf.DefaultMinSubscribe,
		}
		conf.NormalizeBusinessDefaults(&snap)
		raw, _ := json.Marshal(snap)
		if err := db.Create(&SettingPO{Key: conf.SettingsKeySystemConfig, Value: string(raw)}).Error; err != nil {
			return err
		}
	}
	today := time.Now().Format("2006-01-02")
	var priceCnt int64
	if err := db.Model(&AixPricePO{}).Where("effective_date = ?", today).Count(&priceCnt).Error; err != nil {
		return err
	}
	if priceCnt == 0 {
		return db.Create(&AixPricePO{
			Price:         decimal.NewFromInt(1),
			EffectiveDate: today,
			Remark:        "initial",
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
