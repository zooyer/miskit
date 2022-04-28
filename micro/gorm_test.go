package micro

import (
	"encoding/json"
	"testing"

	"gorm.io/gorm"
)

type People struct {
	gorm.Model
	Name string
	Age  int
}

func (People) TableName() string {
	return "people"
}

func TestOpen(t *testing.T) {
	db, err := OpenDB("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatal(err)
	}

	// gorm v2 不用Close了吗？
	// defer db.Close()

	db = db.Debug()

	if err = db.AutoMigrate(new(People)); err != nil {
		t.Fatal(err)
	}

	var p = People{
		Name: "zs",
		Age:  24,
	}

	if err = db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}

	var ps []People
	if err = db.Find(&ps).Error; err != nil {
		t.Fatal(err)
	}

	data, _ := json.Marshal(ps)
	t.Log(string(data))
}
