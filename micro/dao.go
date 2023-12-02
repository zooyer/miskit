package micro

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Model struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at" gorm:"index"`
	DeletedAt int64  `json:"deleted_at" gorm:"index"`
	CreatedID string `json:"created_id" gorm:"index"`
	UpdatedID string `json:"updated_id" gorm:"index"`
	DeletedID string `json:"deleted_id" gorm:"index"`
	CreatedBy string `json:"created_by" gorm:"index"`
	UpdatedBy string `json:"updated_by" gorm:"index"`
	DeletedBy string `json:"deleted_by" gorm:"index"`
}

type CountModel struct {
	CountByYear  int16 `json:"count_by_year" gorm:"index"`  // 按年统计
	CountByMonth int16 `json:"count_by_month" gorm:"index"` // 按月统计
	CountByWeek  int16 `json:"count_by_week" gorm:"index"`  // 按周统计
	CountByDay   int16 `json:"count_by_day" gorm:"index"`   // 按天统计
	CountByHour  int16 `json:"count_by_hour" gorm:"index"`  // 按时统计
}

type Dao struct {
	db      func(ctx context.Context) *gorm.DB
	model   schema.Tabler
	deleted bool
}

type Equal map[string]interface{}

type Update map[string]interface{}

func (d Dao) DB(ctx context.Context) *gorm.DB {
	var db = d.db(ctx).Model(d.model)

	if d.deleted {
		return db
	}

	return db.Scopes(d.undeleted())
}

func (d Dao) QueryDeleted(ok bool) Dao {
	var dao = d
	dao.deleted = ok

	return dao
}

func (d Dao) undeleted() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at = ?", 0)
	}
}

func (d Dao) equal(equal Equal) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if equal == nil {
			return db
		}
		return db.Where(map[string]interface{}(equal))
	}
}

func (d Dao) Equal(ctx context.Context, equal Equal) *gorm.DB {
	return d.DB(ctx).Scopes(d.equal(equal))
}

func (d Dao) Count(ctx context.Context, equal Equal) (count int64, err error) {
	if err = d.Equal(ctx, equal).Count(&count).Error; err != nil {
		return
	}
	return
}

func (d Dao) List(ctx context.Context, query Query, equal Equal, out interface{}) (total int64, err error) {
	if err = d.DB(ctx).Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Scopes(query.ByWhere, query.BySort, query.ByCustom, d.equal(equal)).Count(&total).Error; err != nil {
			return
		}

		if err = tx.Scopes(query.ByQuery, d.equal(equal)).Find(out).Error; err != nil {
			return
		}

		return
	}); err != nil {
		return
	}
	return
}

func (d Dao) First(ctx context.Context, equal Equal, out interface{}) (err error) {
	if err = d.Equal(ctx, equal).First(out).Error; err != nil {
		return
	}
	return
}

func (d Dao) Find(ctx context.Context, equal Equal, out interface{}) (err error) {
	if err = d.Equal(ctx, equal).Find(out).Error; err != nil {
		return
	}
	return
}

func (d Dao) Create(ctx context.Context, equal Equal, value interface{}) (err error) {
	return d.DB(ctx).Transaction(func(tx *gorm.DB) (err error) {
		if len(equal) > 0 {
			var count int64

			if err = tx.Scopes(d.equal(equal)).Count(&count).Error; err != nil {
				return
			}

			if count > 0 {
				return errors.New("record already exists")
			}
		}

		return tx.Create(value).Error
	})
}

func (d Dao) Update(ctx context.Context, equal Equal, update Update) (err error) {
	return d.Equal(ctx, equal).Updates(map[string]interface{}(update)).Error
}

func (d Dao) Delete(ctx context.Context, equal Equal, update Update) (err error) {
	if v := update["updated_id"]; v != nil {
		update["deleted_id"] = v
	}
	if v := update["updated_by"]; v != nil {
		update["deleted_by"] = v
	}
	if v := update["updated_at"]; v != nil {
		update["deleted_at"] = v
	}
	if update["deleted_at"] == nil {
		update["deleted_at"] = time.Now().Unix()
	}

	return d.Update(ctx, equal, update)
}

func NewDao(db func(ctx context.Context) *gorm.DB, model schema.Tabler) Dao {
	return Dao{
		db:    db,
		model: model,
	}
}

func (e Equal) Where() map[string]interface{} {
	return e
}

func (u Update) Updates() map[string]interface{} {
	return u
}
