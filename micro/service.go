package micro

import (
	"context"
	"time"

	"github.com/zooyer/miskit/errors"
)

type Modeler interface {
	SetModelExtra(extra ModelExtra)
}

type Adminer interface {
	CreateModel(now time.Time) ModelExtra
	UpdateModel(model Model, now time.Time) Model
	DeleteModel(now time.Time) Update
	Update(update Update, now time.Time) Update
}

type Sessioner[User Adminer] interface {
	GetUser(ctx context.Context) (user User, err error)
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

type Creator struct {
	Dao   Dao
	Errno int
	Equal Equal
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

type Service[User Adminer, model Modeler] struct {
	Session Sessioner[User]
}

func NewService[User Adminer, Model Modeler](s Sessioner[User]) Service[User, Model] {
	return Service[User, Model]{
		Session: s,
	}
}

func (s *Service[User, Model]) Get(ctx context.Context, get Getter) (m *Model, err error) {
	var model Model

	if err = get.Dao.First(ctx, get.Equal, &model); err != nil {
		return nil, errors.New(get.Errno, err)
	}

	return &model, nil
}

func (s *Service[User, Model]) List(ctx context.Context, list Lister) (result *Result2[Model], err error) {
	var r = Result2[Model]{
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

func (s *Service[User, Model]) Create(ctx context.Context, create Creator) (m *Model, err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	var model Model

	model.SetModelExtra(user.CreateModel(time.Now()))

	if err = create.Dao.Create(ctx, create.Equal, &model); err != nil {
		return nil, errors.New(create.Errno, err)
	}

	return &model, nil
}

func (s *Service[User, Model]) Update(ctx context.Context, update Updater) (err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	user.Update(update.Update, time.Now())

	if err = update.Dao.Update(ctx, update.Equal, update.Update); err != nil {
		return errors.New(update.Errno, err)
	}

	return
}

func (s *Service[User, Model]) Delete(ctx context.Context, delete Deleter) (err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	var update = user.DeleteModel(time.Now())

	if err = delete.Dao.Delete(ctx, delete.Equal, update); err != nil {
		return errors.New(delete.Errno, err)
	}

	return
}
