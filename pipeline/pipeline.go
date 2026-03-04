package pipeline

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tuken/nix/db"
)

type UserContextKey struct{}

var (
	// ...既存のDBKey, LoggerKey...
	UserKey = UserContextKey{}
)

func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		for k, v := range r.Header {
			fmt.Printf("%s: %v\n", k, v)
		}

		cookie, err := r.Cookie("access_token")
		if err != nil {
			// Cookieがなければ未認証のまま
			next.ServeHTTP(w, r)
			return
		}
		// token := cookie.Value
		fmt.Printf("access_token: %#v\n", cookie)

		var user *db.User

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// contextからユーザーを取得するヘルパー
func MustUser(ctx context.Context) *db.User {

	if user, ok := ctx.Value(UserKey).(*db.User); ok {
		return user
	}

	panic("missing user value in context")
}
