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

type Creator struct {
	Dao   Dao
	Errno int
	Equal Equal
	Model Modeler
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

type Service[User Adminer] struct {
	Session Sessioner[User]
}

func NewService(s Sessioner[Adminer]) Service[Adminer] {
	return Service[Adminer]{
		Session: s,
	}
}

func (s *Service[User]) Create(ctx context.Context, create Creator) (err error) {
	user, err := s.Session.GetUser(ctx)
	if err != nil {
		return
	}

	create.Model.SetModelExtra(user.CreateModel(time.Now()))

	if err = create.Dao.Create(ctx, create.Equal, create.Model); err != nil {
		return errors.New(create.Errno, err)
	}

	return
}

func (s *Service[User]) Update(ctx context.Context, update Updater) (err error) {
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

func (s *Service[User]) Delete(ctx context.Context, delete Deleter) (err error) {
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
