package micro

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	ginsession "github.com/gin-contrib/sessions"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zooyer/miskit/imdb"
)

type store struct {
	db      imdb.Conn
	prefix  string
	Codecs  []securecookie.Codec
	Options *sessions.Options
}

type imdbStore struct {
	*store
}

const defaultExpires = int64(20 * time.Minute / time.Second)

func newStore(db imdb.Conn, prefix string, keyPairs ...[]byte) *store {
	return &store{
		db:     db,
		prefix: prefix,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: int(defaultExpires),
		},
	}
}

func (s *store) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *store) New(r *http.Request, name string) (*sessions.Session, error) {
	var (
		err error
		ok  bool
	)

	session := sessions.NewSession(s, name)
	options := *s.Options
	session.Options = &options
	session.IsNew = true

	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			ok, err = s.load(r.Context(), session)
			session.IsNew = err != nil || !ok
		}
	}

	return session, err
}

func (s *store) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		if err := s.delete(r.Context(), session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	} else {
		if session.ID == "" {
			session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := s.save(r.Context(), session); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	}
	return nil
}

func (s *store) load(ctx context.Context, session *sessions.Session) (ok bool, err error) {
	value, err := s.db.Get(ctx, s.prefix+session.ID)
	if err != nil {
		return
	}

	var values = make(map[string]interface{})
	if err = json.Unmarshal([]byte(value), &values); err != nil {
		return
	}

	for key, val := range values {
		session.Values[key] = val
	}

	return true, nil
}

func (s *store) save(ctx context.Context, session *sessions.Session) (err error) {
	var values = make(map[string]interface{})
	for key, value := range session.Values {
		key, ok := key.(string)
		if !ok {
			return errors.New("session key non-string")
		}
		values[key] = value
	}

	data, err := json.Marshal(values)
	if err != nil {
		return
	}
	var expires = defaultExpires
	if session.Options.MaxAge > 0 {
		expires = int64(session.Options.MaxAge)
	}

	if err = s.db.SetEx(ctx, s.prefix+session.ID, string(data), expires); err != nil {
		return
	}

	return
}

func (s *store) delete(ctx context.Context, session *sessions.Session) (err error) {
	if err = s.db.Del(ctx, s.prefix+session.ID); err != nil {
		return
	}

	return
}

func (s *imdbStore) Options(options ginsession.Options) {
	s.store.Options = options.ToGorillaOptions()
}

func NewStore(name, dsn, prefix string, keyPairs ...[]byte) (store ginsession.Store, err error) {
	db, err := imdb.Open(name, dsn)
	if err != nil {
		return
	}

	return NewStoreWithIMDB(db, prefix, keyPairs...), nil
}

func NewStoreWithIMDB(db imdb.Conn, prefix string, keyPairs ...[]byte) ginsession.Store {
	return &imdbStore{
		store: newStore(db, prefix, keyPairs...),
	}
}
