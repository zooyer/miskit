package micro

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Query struct {
	form url.Values // 自定义字段

	Page   int      `form:"page" json:"page,omitempty"`     // 当前页
	Size   int      `form:"size" json:"size,omitempty"`     // 页大小
	Sort   string   `form:"sort" json:"sort,omitempty"`     // 排序
	Cond   string   `form:"cond" json:"cond,omitempty"`     // 条件符号：and、or，用于自定义字段查询
	Omit   []string `form:"omit" json:"omit,omitempty"`     // 忽略字段
	Select []string `form:"select" json:"select,omitempty"` // 选择字段
	Where  string   `form:"where" json:"where,omitempty"`   // 自定义条件SQL
}

type Result struct {
	Query
	Count int         `json:"count"`
	Total int64       `json:"total"`
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

func (q *Query) BySelect(db *gorm.DB) *gorm.DB {
	if len(q.Select) > 0 {
		db = db.Select(q.Select)
	}

	if len(q.Omit) > 0 {
		db = db.Omit(q.Omit...)
	}

	return db
}

func (q *Query) ByLimit(db *gorm.DB) *gorm.DB {
	if q.Page > 0 {
		db = db.Offset((q.Page - 1) * q.Size)
	}

	return db.Limit(q.Size)
}

func (q *Query) BySort(db *gorm.DB) *gorm.DB {
	return db.Order(q.Sort)
}

func (q *Query) ByWhere(db *gorm.DB) *gorm.DB {
	if q.Where != "" {
		db = db.Where(q.Where)
	}
	return db
}

func (q *Query) ByCustom(db *gorm.DB) *gorm.DB {
	var (
		query   string
		args    []interface{}
		isFirst = true
	)

	for key, val := range q.form {
		if omitParams[key] || len(val) == 0 || len(val) == 1 && len(val[0]) == 0 {
			continue
		}

		var op byte // 查询匹配符(^前缀匹配, $后缀匹配, *模糊匹配, =等值匹配, 默认数组匹配)
		if len(val) == 1 && len(val[0]) > 1 {
			switch o := val[0][0]; o {
			case '^', '$', '*', '=', '\\':
				op = o
				val = []string{val[0][1:]}
			}
		}

		// 查询条件
		if isFirst {
			isFirst = false
		} else {
			switch q.Cond {
			case "or":
				query = "OR " + query
			case "and":
				query = "AND " + query
			case "not":
				query = "NOT " + query
			}
		}

		switch op {
		case '^': // 前缀匹配
			query += fmt.Sprintf("%s LIKE ?", key)
			args = append(args, fmt.Sprintf("%s%%", val[0]))
		case '$': // 后缀匹配
			query += fmt.Sprintf("%s LIKE ?", key)
			args = append(args, fmt.Sprintf("%%%s", val[0]))
		case '*': // 模糊匹配
			query += fmt.Sprintf("%s LIKE ?", key)
			args = append(args, fmt.Sprintf("%%%s%%", val[0]))
		case '=': // 等值匹配
			query += fmt.Sprintf("`%s` = ?", key)
			args = append(args, val[0])
		case '\\':
			fallthrough
		default: // 数组匹配 (没有操作符或是数组全部是等值匹配)
			query += fmt.Sprintf("%s IN (?)", key)
			args = append(args, val)
		}
	}

	return db.Where(query, args...)
}

func (q *Query) ByQuery(db *gorm.DB) *gorm.DB {
	return db.Scopes(q.BySelect, q.ByWhere, q.ByCustom, q.ByLimit, q.BySort)
}

func (q *Query) Valid(ctx *gin.Context) (err error) {
	if q.Size == 0 {
		q.Size = 100
	}

	if q.Sort == "" {
		q.Sort = "id DESC"
	}

	if err = ctx.Request.ParseForm(); err != nil {
		return
	}

	q.form = make(url.Values)
	for key, val := range ctx.Request.Form {
		if !omitParams[key] {
			q.form[key] = val
		}
	}

	return
}
