package lang

import (
	"context"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"golang.org/x/text/language"
)

type contextKey struct {
	name string
}

var langCtxKey = &contextKey{"lang"}

// MatchStrings defaults to the first language in this list
// if it does not match anything
var supportedLanguages = language.NewMatcher([]language.Tag{
	language.Hungarian,
	language.English,
})

// Return the user's language from the context
func LangFromContext(ctx context.Context) language.Tag {
	lang, ok := ctx.Value(langCtxKey).(language.Tag)
	assert.True(ok, "Language not found in route context", ctx, lang)

	return lang
}

// Language middleware that puts the language tag in the context
// always should be called after the jwt middleware
func LangMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var langTag language.Tag

		// if the userId is not in the context assume that
		// the request did not come from a user
		userId, ok := jwt.GetUserIDFromContext(ctx)
		if !ok {
			langTag = getLangFromHeader(r)
		} else {
			lang, err := database.PostgreSQL.GetUserPreferredLanguage(database.New(), ctx, userId)
			if err == nil && lang != nil {
				langTag = *lang
			} else {
				langTag = getLangFromHeader(r)
			}
		}

		ctx = context.WithValue(ctx, langCtxKey, langTag)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getLangFromHeader(r *http.Request) language.Tag {
	match, _ := language.MatchStrings(supportedLanguages, r.Header.Get("Accept-Language"))
	tag, _ := match.Base()
	return language.Make(tag.String())
}
