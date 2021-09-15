package micro

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jinzhu/gorm"
)

type Dao struct {
	db    func(ctx context.Context) *gorm.DB
	Model interface{}
	Table string
}

type Where map[string]interface{}

func (d Dao) DB(ctx context.Context) *gorm.DB {
	return d.db(ctx).Table(d.Table)
}

func (d Dao) Equal(ctx context.Context, equal map[string][]interface{}) *gorm.DB {
	db := d.DB(ctx)
	for key, where := range equal {
		switch len(where) {
		case 0:
			continue
		case 1:
			db = db.Where(fmt.Sprintf("%v = ?", key), where[0])
		default:
			db = db.Where(fmt.Sprintf("%v IN (?)", key), where)
		}
	}
	return db
}

func (d Dao) Transaction(ctx context.Context, fn func(tx *gorm.DB) (err error)) (err error) {
	tx := d.DB(ctx).Begin()
	defer func() {
		if err != nil {
			if e := tx.Rollback().Error; e != nil {
				err = fmt.Errorf("transaction error: %w, rollback error: %s", err, e)
			}
		} else {
			err = tx.Commit().Error
		}
	}()
	return fn(tx)
}

func (d Dao) Count(ctx context.Context, where map[string]interface{}) (count int, err error) {
	if err = d.DB(ctx).Where(where).Count(&count).Error; err != nil {
		return
	}
	return
}

func (d Dao) List(ctx context.Context, query Query, form url.Values, where interface{}, out interface{}) (total int, err error) {
	if err = d.Transaction(ctx, func(tx *gorm.DB) (err error) {
		if err = d.DB(ctx).Scopes(ByQuery(d.Model, form)).Where(where).Count(&total).Error; err != nil {
			return
		}

		if err = d.DB(ctx).Scopes(query.ByQuery(d.Model), ByQuery(d.Model, form)).Where(where).Find(out).Error; err != nil {
			return
		}

		return
	}); err != nil {
		return
	}
	return
}

func (d Dao) First(ctx context.Context, out interface{}, equal map[string][]interface{}) (err error) {
	if err = d.Equal(ctx, equal).First(out).Error; err != nil {
		return
	}
	return
}

func (d Dao) Find(ctx context.Context, out interface{}, equal map[string][]interface{}) (err error) {
	if err = d.Equal(ctx, equal).Find(out).Error; err != nil {
		return
	}
	return
}

func (d Dao) Update(ctx context.Context, equal map[string][]interface{}, update map[string]interface{}) (err error) {
	return d.Equal(ctx, equal).Update(update).Error
}

func NewDao(db func(ctx context.Context) *gorm.DB, model interface{}, table string) Dao {
	return Dao{
		db:    db,
		Model: model,
		Table: table,
	}
}

func ByQuery(model interface{}, form url.Values) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		scope := db.NewScope(model)
		for key, val := range form {
			if OmitParam(key) || len(val) == 0 {
				continue
			}

			var hasValue bool // db中有该字段
			if field, ok := scope.FieldByName(key); ok && !field.IsIgnored && len(val) > 0 {
				hasValue = true
			}

			var op byte // 模糊匹配操作符 (等值匹配, ^前缀匹配, $后缀匹配)
			if len(val) == 1 {
				if len(val[0]) > 1 {
					switch o := val[0][0]; o {
					case '\\', '^', '$', '*':
						op = o
						val = []string{val[0][1:]}
					}
				}
				if len(val[0]) == 0 {
					continue
				}
			}

			if hasValue {
				switch op {
				case '^': // 前缀匹配
					db = db.Where(fmt.Sprintf("`%s` LIKE ?", key), fmt.Sprintf("%s%%", val[0]))
				case '$': // 后缀匹配
					db = db.Where(fmt.Sprintf("`%s` LIKE ?", key), fmt.Sprintf("%%%s", val[0]))
				case '*': // 模糊匹配
					db = db.Where(fmt.Sprintf("`%s` LIKE ?", key), fmt.Sprintf("%%%s%%", val[0]))
				default: // 等值匹配 (没有操作符或是数组全部是等值匹配)
					db = db.Where(fmt.Sprintf("`%s` IN (?)", key), val)
				}
			}
		}

		return db
	}
}
