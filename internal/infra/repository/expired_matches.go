package repository

import "gorm.io/gorm"

func ListExpiredMatches(db *gorm.DB, nowUnix int64, limit int) ([]uint, error) {
	var ids []uint
	err := db.Model(&Match{}).
		Where("finished = false").
		Where("turn_deadline_at > 0 AND turn_deadline_at <= ?", nowUnix).
		Limit(limit).Pluck("id", &ids).Error
	return ids, err
}
