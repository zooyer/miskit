package micro

import (
	"context"
	"time"

	"github.com/zooyer/miskit/errors"
	"gorm.io/gorm"
)

type Userinfo interface {
	CreateModel(now time.Time) ModelExtra
	UpdateModel(model Model, now time.Time) Model
	DeleteModel(now time.Time) Update
	Update(update Update, now time.Time) Update
}

type Session[User Userinfo] interface {
	GetUser(ctx context.Context) (user User, err error)
}

type ModelPointer[T any] interface {
	*T
	SetModelExtra(extra ModelExtra)
}

type Getter struct {
	Dao   Dao
	Errno int
	Equal Equal
}

type Lister struct {
	Dao   Dao
	Errno int
	Query Query
}

type Creator[Model any, Pointer ModelPointer[Model]] struct {
	Dao   Dao
	Errno int
	Equal Equal
	Model Model
}

type Updater struct {
	Dao    Dao
	Errno  int
	Equal  Equal
	Update Update
}

type Deleter struct {
	Dao   Dao
	Errno int
	Equal Equal
}

type Service[User Userinfo, Model any, Pointer ModelPointer[Model]] struct {
	Session Session[User]
}

func NewService[User Userinfo, Model any, Pointer ModelPointer[Model]](s Session[User]) Service[User, Model, Pointer] {
	return Service[User, Model, Pointer]{
		Session: s,
	}
}

func (s *Service[User, Model, Pointer]) Get(ctx context.Context, get Getter) (m *Model, err error) {
	var model Model

	if err = get.Dao.First(ctx, get.Equal, &model); err != nil {
		return nil, errors.New(get.Errno, err)
	}

	return &model, nil
}

func (s *Service[User, Model, Pointer]) List(ctx context.Context, list Lister) (result *Result[Model, Pointer], err error) {
	var r = Result[Model, Pointer]{
		Query: list.Query,
		Count: 0,
		Total: 0,
		Data:  nil,
	}

	if r.Total, err = list.Dao.List(ctx, list.Query, nil, &r.Data); err != nil {
		return nil, errors.New(list.Errno, err)
	}

	r.Count = len(r.Data)

	return &r, nil
}

func (s *Service[User, Model, Pointer]) Create(ctx context.Context, create Creator[Model, Pointer]) (m *Model, err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	var model = Pointer(&create.Model)

	model.SetModelExtra(user.CreateModel(time.Now()))

	if err = create.Dao.Create(ctx, create.Equal, model); err != nil {
		return nil, errors.New(create.Errno, err)
	}

	return model, nil
}

func (s *Service[User, Model, Pointer]) Update(ctx context.Context, update Updater) (m *Model, err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	user.Update(update.Update, time.Now())

	var model Model

	if err = update.Dao.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Scopes(update.Dao.equal(update.Equal)).Updates(update.Update).First(&model).Error
	}); err != nil {
		return nil, errors.New(update.Errno, err)
	}

	return &model, nil
}

func (s *Service[User, Model, Pointer]) Delete(ctx context.Context, delete Deleter) (m *Model, err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	var update = user.DeleteModel(time.Now())

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

	var model Model

	if err = delete.Dao.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Scopes(delete.Dao.equal(delete.Equal)).Updates(update).First(&model).Error
	}); err != nil {
		return nil, errors.New(delete.Errno, err)
	}

	return &model, nil
}
