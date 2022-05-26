package micro

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Model struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at" sql:"index"`
	CreatedID int64  `json:"created_id"`
	UpdatedID int64  `json:"updated_id"`
	DeletedID int64  `json:"deleted_id"`
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	DeletedBy string `json:"deleted_by"`
}

type Dao struct {
	db      func(ctx context.Context) *gorm.DB
	model   schema.Tabler
	deleted bool
}

type Equal map[string]interface{}

type Update map[string]interface{}

type Include map[string][]interface{}

func (d Dao) DB(ctx context.Context) *gorm.DB {
	var db = d.db(ctx).Model(d.model)

	if d.deleted {
		return db
	}

	return db.Scopes(d.undeleted())
}

func (d Dao) QueryDeleted(ok bool) {
	d.deleted = true
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

func (d Dao) include(include Include) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for key, where := range include {
			db = db.Where(fmt.Sprintf("%v IN (?)", key), where)
		}
		return db
	}
}

func (d Dao) Equal(ctx context.Context, equal Equal) *gorm.DB {
	return d.DB(ctx).Scopes(d.equal(equal))
}

func (d Dao) Include(ctx context.Context, include Include) *gorm.DB {
	return d.DB(ctx).Scopes(d.include(include))
}

func (d Dao) Count(ctx context.Context, equal Equal) (count int64, err error) {
	if err = d.Equal(ctx, equal).Count(&count).Error; err != nil {
		return
	}
	return
}

func (d Dao) List(ctx context.Context, query Query, include Include, out interface{}) (total int64, err error) {
	if err = d.DB(ctx).Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Scopes(query.ByWhere, query.BySort, query.ByCustom, d.include(include)).Count(&total).Error; err != nil {
			return
		}

		if err = tx.Scopes(query.ByQuery, d.include(include)).Find(out).Error; err != nil {
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

func (d Dao) Delete(ctx context.Context, equal Equal) (err error) {
	return d.Update(ctx, equal, Update{
		"deleted_at": time.Now(),
	})
}

func NewDao(db func(ctx context.Context) *gorm.DB, model schema.Tabler) Dao {
	return Dao{
		db:    db,
		model: model,
	}
}
