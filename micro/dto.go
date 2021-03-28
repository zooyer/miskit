package micro

import (
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

type Query struct {
	Page   int      `form:"page" json:"page,omitempty"`
	Size   int      `form:"size" json:"size,omitempty"`
	Sort   string   `form:"sort" json:"sort,omitempty"`
	Omit   []string `form:"omit" json:"omit,omitempty"`
	Select []string `form:"select" json:"select,omitempty"`
	Where  string   `form:"where" json:"where,omitempty"`
}

type Result struct {
	Query
	Count int         `json:"count"`
	Total int         `json:"total"`
	Data  interface{} `json:"data"`
}

var omitParams = make(map[string]bool)

func init() {
	t := reflect.TypeOf(Query{})
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("form")
		if tag == "" {
			tag = t.Field(i).Tag.Get("json")
			tag = strings.Split(tag, ",")[0]
		}
		omitParams[tag] = true
	}
}

func OmitParam(key string) bool {
	return omitParams[key]
}

func (q Query) BySelect(model interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// select field
		var selected = make(map[string]bool)
		for _, s := range q.Select {
			selected[s] = true
		}

		var scope = db.NewScope(model)
		// omit field
		if len(q.Omit) > 0 {
			if len(selected) == 0 {
				for _, field := range scope.Fields() {
					if !field.IsIgnored {
						selected[field.DBName] = true
					}
				}
			}
			for _, omit := range q.Omit {
				delete(selected, omit)
			}
		}

		// db select field
		if len(selected) > 0 {
			var fields = make([]string, 0, len(selected))
			for field := range selected {
				if _, ok := scope.FieldByName(field); ok {
					fields = append(fields, scope.Quote(field))
				}
			}
			db = db.Select(fields)
		}

		return db
	}
}

func (q Query) ByLimit(db *gorm.DB) *gorm.DB {
	if q.Size > 0 {
		db = db.Limit(q.Size)
		if q.Page > 0 {
			db = db.Offset((q.Page - 1) * q.Size)
		}
	} else {
		db = db.Limit(1000)
	}

	return db
}

func (q Query) BySort(db *gorm.DB) *gorm.DB {
	if q.Sort != "" {
		db = db.Order(q.Sort)
	} else {
		db = db.Order("updated_at DESC")
	}
	return db
}

func (q Query) ByQuery(model interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = q.BySelect(model)(db)
		db = q.ByLimit(db)
		db = q.BySort(db)
		return db
	}
}

func (q Query) ByWhere(db *gorm.DB) *gorm.DB {
	if q.Where != "" {
		db = db.Where(q.Where)
	}
	return db
}
