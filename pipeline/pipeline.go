package pipeline

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oklog/ulid/v2"
	"github.com/tuken/nix/db"
	"github.com/tuken/nix/logger"
	"gorm.io/gorm"
)

type contextKey struct{ string }

var (
	DBKey     = contextKey{"database"}
	LoggerKey = contextKey{"logger"}
	UserKey   = contextKey{"user"}
)

func MustDB(ctx context.Context) *gorm.DB {

	if db, ok := ctx.Value(DBKey).(*gorm.DB); ok {
		return db
	}

	panic("missing db value in context")
}

func MustLogger(ctx context.Context) logger.Logger {

	if log, ok := ctx.Value(LoggerKey).(logger.Logger); ok {
		return log
	}

	panic("missing logger value in context")
}

func MustUser(ctx context.Context) *db.User {

	if user, ok := ctx.Value(UserKey).(*db.User); ok {
		return user
	}

	panic("missing user value in context")
}

type Preprocessor struct {
	DB  *gorm.DB
	Log logger.Logger
}

func (p *Preprocessor) Pipeline(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		for k, v := range r.Header {
			fmt.Printf("%s: %v\n", k, v)
		}

		requestID := ulid.Make().String()

		ctx := r.Context()
		d := p.DB.WithContext(ctx)

		l := p.Log.With("rid", requestID)

		ctx = context.WithValue(ctx, DBKey, d)
		ctx = context.WithValue(ctx, LoggerKey, l)

		ref := r.Referer()
		orgRoot := r.Header.Get("Origin") + "/"

		if ref != orgRoot {

			cookie, err := r.Cookie("session")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			fmt.Printf("session: %#v\n", cookie)

			user := db.User{}
			if err := d.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields").Joins("INNER JOIN sessions s ON (s.data -> '$.user.id') = users.id").Where("s.session_id = ?", cookie.Value).Last(&user).Error; err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, UserKey, &user)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
