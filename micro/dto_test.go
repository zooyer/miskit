package micro

import (
	"encoding/json"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
)

func TestQuery_ByQuery(t *testing.T) {
	var _ sqlite3.SQLiteDriver
	type Stu struct {
		gorm.Model
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Level int    `json:"level"`
	}
	var query = Query{
		Page:   0,
		Size:   0,
		Sort:   "",
		Omit:   []string{"level"},
		Select: []string{"name", "age"},
		Where:  "",
	}

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db = db.Debug()
	db.SingularTable(true)
	db.AutoMigrate(&Stu{})

	db.Create(&Stu{Name: "xiaozhang", Age: 28, Level: 3})
	db.Create(&Stu{Name: "xiaoli", Age: 19, Level: 32})
	db.Create(&Stu{Name: "xiaowang", Age: 22, Level: 5})
	db.Create(&Stu{Name: "xiaohong", Age: 18, Level: 1})

	var stu []Stu
	if err := db.Scopes(query.ByQuery(&Stu{})).Find(&stu).Error; err != nil {
		t.Fatal(err)
	}

	data, err := json.MarshalIndent(stu, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
