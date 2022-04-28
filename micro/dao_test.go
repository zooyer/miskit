package micro

import (
	"context"
	"gorm.io/gorm"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testModel struct {
	Model
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (t testModel) TableName() string {
	return "test"
}

func TestNewDao(t *testing.T) {
	var ctx = context.Background()

	db, err := OpenDB("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatal(err)
	}

	db = db.Debug()

	if err = db.AutoMigrate(new(testModel)); err != nil {
		t.Fatal(err)
	}

	getDB := func(ctx context.Context) *gorm.DB {
		return db.WithContext(ctx)
	}

	dao := NewDao(getDB, new(testModel))

	var test = testModel{
		Name: "dog",
	}

	if err = dao.Create(ctx, nil, &test); err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", test)

	test.ID = 0
	if err = dao.Create(ctx, nil, &test); err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", test)

	count, err := dao.Count(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(2), count)
}
