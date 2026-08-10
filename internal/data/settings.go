package data

import (
	"context"

	"backend/internal/biz"

	"gorm.io/gorm"
)

type settingsRepo struct {
	data *Data
}

func NewSettingsRepo(data *Data) biz.SettingsRepo {
	return &settingsRepo{data: data}
}

func (r *settingsRepo) Get(ctx context.Context, key string) (string, error) {
	var po SettingPO
	err := r.data.db.WithContext(ctx).Where("`key` = ?", key).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return po.Value, nil
}

func (r *settingsRepo) Set(ctx context.Context, key, value string) error {
	var po SettingPO
	err := r.data.db.WithContext(ctx).Where("`key` = ?", key).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return r.data.db.WithContext(ctx).Create(&SettingPO{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&po).Update("value", value).Error
}

func (r *settingsRepo) GetAll(ctx context.Context) (map[string]string, error) {
	var list []SettingPO
	if err := r.data.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, item := range list {
		result[item.Key] = item.Value
	}
	return result, nil
}
